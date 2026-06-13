package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	internaloidc "github.com/haierkeys/fast-note-sync-service/internal/oidc"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

// OIDCService authenticates WebGUI users with an external OIDC provider.
type OIDCService interface {
	Authenticate(ctx context.Context, claims internaloidc.Claims, clientIP, clientType, userAgent string) (*dto.UserDTO, error)
}

type OIDCUserMappingConfig struct {
	SubjectClaim     string
	EmailClaim       string
	UsernameClaim    string
	DisplayNameClaim string
}

type OIDCServiceConfig struct {
	AutoRegister bool
	Issuer       string
	UserMapping  OIDCUserMappingConfig
}

type loginTokenIssuer interface {
	CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error)
}

type oidcService struct {
	userRepo     domain.UserRepository
	identityRepo domain.OIDCIdentityRepository
	tokenService loginTokenIssuer
	config       OIDCServiceConfig
}

func NewOIDCService(userRepo domain.UserRepository, identityRepo domain.OIDCIdentityRepository, tokenService loginTokenIssuer, config OIDCServiceConfig) OIDCService {
	return &oidcService{
		userRepo:     userRepo,
		identityRepo: identityRepo,
		tokenService: tokenService,
		config:       config,
	}
}

func (s *oidcService) Authenticate(ctx context.Context, claims internaloidc.Claims, clientIP, clientType, userAgent string) (*dto.UserDTO, error) {
	claims = s.mappedClaims(claims)
	subject := strings.TrimSpace(claims.Subject)
	if subject == "" {
		return nil, code.ErrorUserLoginFailed.WithDetails("oidc subject is empty")
	}

	issuer := strings.TrimSpace(s.config.Issuer)
	if issuer == "" {
		return nil, code.ErrorUserLoginFailed.WithDetails("oidc issuer is empty")
	}

	user, err := s.findBoundUser(ctx, issuer, subject)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = s.findOrCreateUser(ctx, claims)
		if err != nil {
			return nil, err
		}
		if _, err := s.identityRepo.Create(ctx, &domain.OIDCIdentity{
			UID:      user.UID,
			Issuer:   issuer,
			Subject:  subject,
			Email:    claims.Email,
			Username: claims.Username,
		}); err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
	}

	if clientType == "" {
		clientType = "WebGUI"
	}
	token, tokenStr, err := s.tokenService.CreateForLogin(ctx, user.UID, clientType, clientIP, userAgent)
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	return &dto.UserDTO{
		UID:       user.UID,
		Email:     user.Email,
		Username:  user.Username,
		Avatar:    user.Avatar,
		Token:     tokenStr,
		TokenID:   token.ID,
		UpdatedAt: timex.Time(user.UpdatedAt),
		CreatedAt: timex.Time(user.CreatedAt),
	}, nil
}

func (s *oidcService) mappedClaims(claims internaloidc.Claims) internaloidc.Claims {
	mapping := s.config.UserMapping
	if claims.Raw == nil {
		return claims
	}
	if mapping.SubjectClaim != "" {
		claims.Subject = rawStringClaim(claims.Raw, mapping.SubjectClaim)
	}
	if mapping.EmailClaim != "" {
		claims.Email = rawStringClaim(claims.Raw, mapping.EmailClaim)
	}
	if mapping.UsernameClaim != "" {
		claims.Username = rawStringClaim(claims.Raw, mapping.UsernameClaim)
	}
	if mapping.DisplayNameClaim != "" {
		claims.DisplayName = rawStringClaim(claims.Raw, mapping.DisplayNameClaim)
	}
	return claims
}

func rawStringClaim(claims map[string]interface{}, name string) string {
	value, _ := claims[name].(string)
	return strings.TrimSpace(value)
}

func (s *oidcService) findBoundUser(ctx context.Context, issuer, subject string) (*domain.User, error) {
	identity, err := s.identityRepo.GetByIssuerSubject(ctx, issuer, subject)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	user, err := s.userRepo.GetByUID(ctx, identity.UID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorUserNotFound
	}
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return user, nil
}

func (s *oidcService) findOrCreateUser(ctx context.Context, claims internaloidc.Claims) (*domain.User, error) {
	email := strings.TrimSpace(claims.Email)
	if email != "" {
		user, err := s.userRepo.GetByEmail(ctx, email)
		if err == nil {
			return user, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
	}

	if !s.config.AutoRegister {
		return nil, code.ErrorUserNotFound
	}
	if email == "" || !util.IsValidEmail(email) {
		return nil, code.ErrorUserRegister.WithDetails("oidc email is required for auto registration")
	}

	username, err := s.availableUsername(ctx, usernameCandidates(claims)...)
	if err != nil {
		return nil, err
	}
	password, err := util.GeneratePasswordHash(util.GetRandomString(32))
	if err != nil {
		return nil, code.ErrorPasswordNotValid
	}

	user, err := s.userRepo.Create(ctx, &domain.User{
		Email:    email,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, code.ErrorUserRegister.WithDetails(err.Error())
	}
	return user, nil
}

func usernameCandidates(claims internaloidc.Claims) []string {
	candidates := []string{
		claims.Username,
		claims.DisplayName,
		emailLocalPart(claims.Email),
	}
	if claims.Subject != "" {
		candidates = append(candidates, "oidc_"+claims.Subject)
	}
	return candidates
}

func (s *oidcService) availableUsername(ctx context.Context, candidates ...string) (string, error) {
	base := ""
	for _, candidate := range candidates {
		base = normalizeUsername(candidate)
		if base != "" {
			break
		}
	}
	if base == "" {
		base = "oidc_user"
	}

	for i := 0; i < 100; i++ {
		username := base
		if i > 0 {
			username = truncateUsername(base, 17) + fmt.Sprintf("%03d", i)
		}
		_, err := s.userRepo.GetByUsername(ctx, username)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return username, nil
		}
		if err != nil {
			return "", code.ErrorDBQuery.WithDetails(err.Error())
		}
	}
	return "", code.ErrorUserAlreadyExists
}

var oidcUsernameInvalidChars = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func normalizeUsername(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "-", "_")
	value = oidcUsernameInvalidChars.ReplaceAllString(value, "_")
	if len(value) > 20 {
		value = value[:20]
	}
	if util.IsValidUsername(value) {
		return value
	}
	return ""
}

func emailLocalPart(email string) string {
	local, _, ok := strings.Cut(strings.TrimSpace(email), "@")
	if !ok {
		return ""
	}
	return local
}

func truncateUsername(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
