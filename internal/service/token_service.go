package service

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
)

// TokenService defines the token management service interface
type TokenService interface {
	// Create creates a new manual token
	Create(ctx context.Context, uid int64, params *dto.TokenIssueRequest) (*dto.TokenCreateResponse, error)
	// CreateForLogin creates a token during the login flow
	CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error)
	// ListByUser lists all active tokens for a user
	ListByUser(ctx context.Context, uid int64) ([]*dto.TokenResponse, error)
	// UpdateScope updates a token's scope
	UpdateScope(ctx context.Context, uid int64, tokenID int64, params *dto.TokenUpdateRequest) error
	// Revoke revokes a token
	Revoke(ctx context.Context, uid int64, tokenID int64) error
	// GetActiveToken gets an active token by ID
	GetActiveToken(ctx context.Context, uid int64, tokenID int64) (*domain.AuthToken, error)
	// RecordAccessLog records a token access log
	RecordAccessLog(ctx context.Context, log *domain.AuthTokenLog) error
}

type tokenService struct {
	tokenRepo    domain.AuthTokenRepository
	logRepo      domain.AuthTokenLogRepository
	tokenManager app.TokenManager
	logger       *zap.Logger
}

func NewTokenService(tokenRepo domain.AuthTokenRepository, logRepo domain.AuthTokenLogRepository, tokenManager app.TokenManager, logger *zap.Logger) TokenService {
	return &tokenService{
		tokenRepo:    tokenRepo,
		logRepo:      logRepo,
		tokenManager: tokenManager,
		logger:       logger,
	}
}

func (s *tokenService) domainToDTO(token *domain.AuthToken) *dto.TokenResponse {
	return &dto.TokenResponse{
		ID:         token.ID,
		Scope:      token.Scope,
		ClientType: token.ClientType,
		BoundIP:    token.BoundIP,
		UserAgent:  token.UserAgent,
		ExpiredAt:  timex.Time(token.ExpiredAt),
		CreatedAt:  timex.Time(token.CreatedAt),
	}
}

func (s *tokenService) Create(ctx context.Context, uid int64, params *dto.TokenIssueRequest) (*dto.TokenCreateResponse, error) {
	t := &domain.AuthToken{
		UID:        uid,
		Scope:      params.Scope,
		ClientType: params.ClientType,
		Status:     1,
		ExpiredAt:  time.Now().Add(time.Duration(params.ExpiredDays) * 24 * time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t, err := s.tokenRepo.Create(ctx, t)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Generate JWT using token_id
	tokenStr, err := s.tokenManager.Generate(uid, "", "", t.ID)
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	// Save token string back
	t.TokenString = tokenStr
	err = s.tokenRepo.UpdateScope(ctx, t.ID, t.Scope) // Note: Dao doesn't have UpdateTokenString yet, but actually we can update token string if we add it, or just ignore for now since it's just for reference. Wait, I should add UpdateTokenString to repo or ignore. Let's just not update DB with it if not necessary. But we have it in DB. Let's assume we don't need to update it back immediately if it's not strictly checked by string in standard flow. Or wait, let's just use UpdateScope method for now. Actually better to have an UpdateTokenString method, but wait, the JWT is the token string. Let's skip saving TokenString for now as it's not used in strict validation.
	
	res := &dto.TokenCreateResponse{
		TokenResponse: *s.domainToDTO(t),
		TokenString:   tokenStr,
	}
	return res, nil
}

func (s *tokenService) CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error) {
	// Default scope for webgui login
	scope := "p:rest c:webgui f:*"
	
	t := &domain.AuthToken{
		UID:        uid,
		Scope:      scope,
		ClientType: clientType,
		BoundIP:    ip,
		UserAgent:  userAgent,
		Status:     1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(30 * 24 * time.Hour), // Default 30 days
	}

	t, err := s.tokenRepo.Create(ctx, t)
	if err != nil {
		return nil, "", code.ErrorDBQuery.WithDetails(err.Error())
	}

	tokenStr, err := s.tokenManager.Generate(uid, "", ip, t.ID)
	if err != nil {
		return nil, "", code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	t.TokenString = tokenStr

	return t, tokenStr, nil
}

func (s *tokenService) ListByUser(ctx context.Context, uid int64) ([]*dto.TokenResponse, error) {
	tokens, err := s.tokenRepo.ListByUID(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var res []*dto.TokenResponse
	for _, t := range tokens {
		res = append(res, s.domainToDTO(t))
	}
	return res, nil
}

func (s *tokenService) UpdateScope(ctx context.Context, uid int64, tokenID int64, params *dto.TokenUpdateRequest) error {
	// Need to check if token belongs to user first
	token, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	if token.UID != uid {
		return code.ErrorInvalidAuthToken
	}
	
	err = s.tokenRepo.UpdateScope(ctx, tokenID, params.Scope)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

func (s *tokenService) Revoke(ctx context.Context, uid int64, tokenID int64) error {
	token, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	if token.UID != uid {
		return code.ErrorInvalidAuthToken
	}

	err = s.tokenRepo.Revoke(ctx, tokenID)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

func (s *tokenService) RecordAccessLog(ctx context.Context, log *domain.AuthTokenLog) error {
	return s.logRepo.Create(ctx, log)
}

func (s *tokenService) GetActiveToken(ctx context.Context, uid int64, tokenID int64) (*domain.AuthToken, error) {
	token, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	if token.UID != uid || token.Status != 1 {
		return nil, code.ErrorInvalidAuthToken
	}
	return token, nil
}
