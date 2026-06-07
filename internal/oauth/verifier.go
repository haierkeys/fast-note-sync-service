package oauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTVerifierConfig struct {
	Issuer         string
	Audience       string
	JWKSURL        string
	RequiredScopes []string
	HTTPClient     *http.Client
}

type VerifiedClaims struct {
	Subject string
	Scopes  []string
	Raw     map[string]interface{}
}

type JWTVerifier struct {
	config JWTVerifierConfig
}

func NewJWTVerifier(config JWTVerifierConfig) *JWTVerifier {
	return &JWTVerifier{config: config}
}

func (v *JWTVerifier) Verify(ctx context.Context, tokenString string) (*VerifiedClaims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, fmt.Errorf("%w: token is empty", ErrInvalidToken)
	}
	if strings.TrimSpace(v.config.JWKSURL) == "" {
		return nil, fmt.Errorf("%w: jwks url is required", ErrConfig)
	}

	claims := jwt.MapClaims{}
	parserOptions := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	}
	if v.config.Issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(v.config.Issuer))
	}
	if v.config.Audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(v.config.Audience))
	}

	parser := jwt.NewParser(parserOptions...)
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %s", ErrInvalidToken, token.Method.Alg())
		}

		kid, _ := token.Header["kid"].(string)
		if strings.TrimSpace(kid) == "" {
			return nil, fmt.Errorf("%w: missing kid", ErrInvalidToken)
		}

		key, err := v.publicKeyForKID(ctx, kid)
		if err != nil {
			return nil, err
		}
		return key, nil
	})
	if err != nil {
		if errors.Is(err, ErrConfig) || errors.Is(err, ErrInsufficientScope) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if token == nil || !token.Valid {
		return nil, fmt.Errorf("%w: jwt validation failed", ErrInvalidToken)
	}

	scopes := extractScopes(claims)
	if err := requireScopes(scopes, v.config.RequiredScopes); err != nil {
		return nil, err
	}

	subject, _ := claims["sub"].(string)
	return &VerifiedClaims{
		Subject: strings.TrimSpace(subject),
		Scopes:  scopes,
		Raw:     mapClaimsToRaw(claims),
	}, nil
}

func (v *JWTVerifier) publicKeyForKID(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.config.JWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfig, err)
	}

	client := v.config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: jwks status %d", ErrInvalidToken, resp.StatusCode)
	}

	var set jwksSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	for _, key := range set.Keys {
		if key.KID != kid {
			continue
		}
		return key.rsaPublicKey()
	}

	return nil, fmt.Errorf("%w: unknown kid %q", ErrInvalidToken, kid)
}

type jwksSet struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KTY string `json:"kty"`
	KID string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (k jwk) rsaPublicKey() (*rsa.PublicKey, error) {
	if !strings.EqualFold(k.KTY, "RSA") {
		return nil, fmt.Errorf("%w: unsupported jwk kty %q", ErrInvalidToken, k.KTY)
	}
	if k.Use != "" && !strings.EqualFold(k.Use, "sig") {
		return nil, fmt.Errorf("%w: unsupported jwk use %q", ErrInvalidToken, k.Use)
	}
	if k.Alg != "" && !strings.HasPrefix(strings.ToUpper(k.Alg), "RS") {
		return nil, fmt.Errorf("%w: unsupported jwk alg %q", ErrInvalidToken, k.Alg)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid jwk modulus", ErrInvalidToken)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid jwk exponent", ErrInvalidToken)
	}

	exponent := int(new(big.Int).SetBytes(eBytes).Int64())
	if exponent <= 0 {
		return nil, fmt.Errorf("%w: invalid jwk exponent", ErrInvalidToken)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: exponent,
	}, nil
}

func extractScopes(claims jwt.MapClaims) []string {
	var scopes []string
	appendScope := func(scope string) {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	appendClaimScopes := func(value interface{}) {
		switch value := value.(type) {
		case string:
			for _, scope := range strings.Fields(value) {
				appendScope(scope)
			}
		case []string:
			for _, scope := range value {
				appendScope(scope)
			}
		case []interface{}:
			for _, item := range value {
				if scope, ok := item.(string); ok {
					appendScope(scope)
				}
			}
		}
	}

	appendClaimScopes(claims["scope"])
	appendClaimScopes(claims["scp"])
	appendClaimScopes(claims["permissions"])

	return scopes
}

func requireScopes(scopes []string, required []string) error {
	if len(required) == 0 {
		return nil
	}

	granted := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		granted[scope] = true
	}

	for _, scope := range required {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if !granted[scope] {
			return fmt.Errorf("%w: missing %q", ErrInsufficientScope, scope)
		}
	}

	return nil
}

func mapClaimsToRaw(claims jwt.MapClaims) map[string]interface{} {
	raw := make(map[string]interface{}, len(claims))
	for key, value := range claims {
		raw[key] = value
	}
	return raw
}
