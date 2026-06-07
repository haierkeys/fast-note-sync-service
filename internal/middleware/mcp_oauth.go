package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	internaloauth "github.com/haierkeys/fast-note-sync-service/internal/oauth"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

type MCPOAuthVerifier interface {
	VerifyBearerToken(ctx context.Context, token string, clientType string) (*MCPOAuthClaims, error)
}

type MCPOAuthScopeMapper func(*MCPOAuthClaims) string

type MCPOAuthClaims struct {
	UID           int64
	UserID        int64
	Scope         string
	FNSScope      string
	OAuthScopes   []string
	ClientID      string
	ClientName    string
	ClientVersion string
}

type MCPMixedAuthConfig struct {
	OAuthEnabled         bool
	Resource             string
	AllowStaticFNSToken  bool
	Verifier             MCPOAuthVerifier
	ScopeMapper          MCPOAuthScopeMapper
	DefaultClient        string
	DefaultClientName    string
	DefaultClientVersion string
}

func MCPOAuthWithConfig(oauthCfg config.OAuthConfig, secretKey string, tokenService service.TokenService, userRepo domain.UserRepository) gin.HandlerFunc {
	oauthCfg.Normalize()
	return MCPMixedAuthWithConfig(secretKey, tokenService, MCPMixedAuthConfig{
		OAuthEnabled:         oauthCfg.Enabled,
		Resource:             oauthCfg.Resource,
		AllowStaticFNSToken:  oauthCfg.AllowStaticFNSToken,
		Verifier:             newOAuthVerifierAdapter(oauthCfg, userRepo),
		DefaultClient:        oauthCfg.DefaultClient,
		DefaultClientName:    oauthCfg.DefaultClientName,
		DefaultClientVersion: oauthCfg.DefaultClientVersion,
	})
}

func MCPMixedAuthWithConfig(secretKey string, tokenService service.TokenService, cfg MCPMixedAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.OAuthEnabled {
			UserAuthTokenWithConfig(secretKey, tokenService)(c)
			return
		}

		if cfg.AllowStaticFNSToken && ExtractUserAuthToken(c) != "" {
			if user, scope, vaults, err := AuthenticateUserToken(c, secretKey, tokenService); err == nil {
				c.Set("user_token", user)
				c.Set("scope", scope)
				c.Set("vaults", vaults)
				c.Next()
				return
			}
		}

		token := extractOAuthBearerToken(c)
		if token == "" || cfg.Verifier == nil {
			log.Printf("mcp oauth rejected: reason=missing_bearer_token path=%s client=%q verifier_configured=%t", c.Request.URL.Path, c.GetHeader("X-Client"), cfg.Verifier != nil)
			writeOAuthChallenge(c, http.StatusUnauthorized, cfg.Resource, "invalid_token")
			return
		}

		ensureMCPClientHeaders(c, cfg, nil)
		claims, err := cfg.Verifier.VerifyBearerToken(c.Request.Context(), token, c.GetHeader("X-Client"))
		if err != nil || claims == nil {
			log.Printf("mcp oauth rejected: reason=verify_failed path=%s client=%q insufficient_scope=%t err=%q", c.Request.URL.Path, c.GetHeader("X-Client"), errors.Is(err, internaloauth.ErrInsufficientScope), err)
			if errors.Is(err, internaloauth.ErrInsufficientScope) {
				writeOAuthChallenge(c, http.StatusForbidden, cfg.Resource, "insufficient_scope")
				return
			}
			writeOAuthChallenge(c, http.StatusUnauthorized, cfg.Resource, "invalid_token")
			return
		}

		uid := claims.UID
		if uid == 0 {
			uid = claims.UserID
		}
		scope := claims.FNSScope
		if cfg.ScopeMapper != nil {
			scope = cfg.ScopeMapper(claims)
		}
		if scope == "" {
			scope = claims.Scope
		}

		ensureMCPClientHeaders(c, cfg, claims)
		if !pkgapp.VerifyPermissions(scope, "mcp", c.GetHeader("X-Client"), "") {
			log.Printf("mcp oauth rejected: reason=fns_permission_failed path=%s uid=%d client=%q fns_scope=%q oauth_scopes=%q", c.Request.URL.Path, uid, c.GetHeader("X-Client"), scope, strings.Join(claims.OAuthScopes, " "))
			writeOAuthChallenge(c, http.StatusForbidden, cfg.Resource, "insufficient_scope")
			return
		}

		c.Set("user_token", &pkgapp.UserEntity{UID: uid})
		c.Set("scope", scope)
		c.Next()
	}
}

func extractOAuthBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

func ensureMCPClientHeaders(c *gin.Context, cfg MCPMixedAuthConfig, claims *MCPOAuthClaims) {
	var claimsClientID, claimsClientName, claimsClientVersion string
	if claims != nil {
		claimsClientID = claims.ClientID
		claimsClientName = claims.ClientName
		claimsClientVersion = claims.ClientVersion
	}
	clientType := firstNonEmpty(c.GetHeader("X-Client"), claimsClientID, cfg.DefaultClient, "MCP")
	clientName := firstNonEmpty(c.GetHeader("X-Client-Name"), claimsClientName, cfg.DefaultClientName, clientType)
	clientVersion := firstNonEmpty(c.GetHeader("X-Client-Version"), claimsClientVersion, cfg.DefaultClientVersion)

	c.Request.Header.Set("X-Client", clientType)
	c.Request.Header.Set("X-Client-Name", clientName)
	if clientVersion != "" {
		c.Request.Header.Set("X-Client-Version", clientVersion)
	}
}

func writeOAuthChallenge(c *gin.Context, status int, resource string, authErr string) {
	challenge := `Bearer`
	if resource != "" {
		challenge += ` resource_metadata="` + escapeAuthParam(oauthProtectedResourceMetadataURL(resource)) + `"`
		challenge += ` resource="` + escapeAuthParam(resource) + `"`
	}
	if authErr != "" {
		challenge += `, error="` + escapeAuthParam(authErr) + `"`
	}
	c.Header("WWW-Authenticate", challenge)
	c.AbortWithStatusJSON(status, gin.H{
		"error": authErr,
		"_meta": gin.H{
			"mcp/www_authenticate": challenge,
		},
	})
}

func oauthProtectedResourceMetadataURL(resource string) string {
	u, err := url.Parse(resource)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return resource
	}
	resourcePath := strings.TrimRight(u.Path, "/")
	u.Path = "/.well-known/oauth-protected-resource" + resourcePath
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func escapeAuthParam(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

type oauthVerifierAdapter struct {
	jwtVerifier   *internaloauth.JWTVerifier
	subjectMapper *internaloauth.SubjectMapper
	client        string
	defaultScope  string
}

func newOAuthVerifierAdapter(cfg config.OAuthConfig, userRepo domain.UserRepository) *oauthVerifierAdapter {
	audience := cfg.Resource
	if len(cfg.Audience) > 0 {
		audience = cfg.Audience[0]
	}
	return &oauthVerifierAdapter{
		jwtVerifier: internaloauth.NewJWTVerifier(internaloauth.JWTVerifierConfig{
			Issuer:         cfg.Issuer,
			Audience:       audience,
			JWKSURL:        cfg.JWKSURL,
			RequiredScopes: requiredOAuthScopes(cfg),
		}),
		subjectMapper: internaloauth.NewSubjectMapper(userRepo, internaloauth.SubjectMapperConfig{
			Mode:       cfg.SubjectMapping.Mode,
			EmailClaim: cfg.SubjectMapping.Claim,
			FixedUID:   cfg.SubjectMapping.FixedUID,
		}),
		client:       cfg.DefaultClient,
		defaultScope: cfg.DefaultFNSScope,
	}
}

func requiredOAuthScopes(cfg config.OAuthConfig) []string {
	if strings.TrimSpace(cfg.DefaultFNSScope) != "" {
		return nil
	}
	return cfg.RequiredScopes
}

func (v *oauthVerifierAdapter) VerifyBearerToken(ctx context.Context, token string, clientType string) (*MCPOAuthClaims, error) {
	claims, err := v.jwtVerifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}
	uid, err := v.subjectMapper.Map(ctx, claims.Raw)
	if err != nil {
		return nil, err
	}
	fnsScope := v.defaultScope
	if fnsScope == "" {
		client := firstNonEmpty(clientType, v.client)
		fnsScope, err = internaloauth.MapOAuthScopesToFNS(client, claims.Scopes)
		if err != nil {
			return nil, err
		}
	}
	return &MCPOAuthClaims{
		UID:         uid,
		Scope:       fnsScope,
		FNSScope:    fnsScope,
		OAuthScopes: claims.Scopes,
	}, nil
}
