package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TokenService defines the token management service interface
type TokenService interface {
	// Create creates a new manual token
	Create(ctx context.Context, uid int64, params *dto.TokenIssueRequest) (*dto.TokenCreateResponse, error)
	// CreateForLogin creates a token during the login flow
	CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error)
	// ListByUser lists all active tokens for a user
	ListByUser(ctx context.Context, uid int64) ([]*dto.TokenResponse, error)
	// Update updates a token's properties
	Update(ctx context.Context, uid int64, tokenID int64, params *dto.TokenUpdateRequest) error
	// Revoke revokes a token
	Revoke(ctx context.Context, uid int64, tokenID int64) error
	// Rotate rotates a token (generates new JWT and invalidates old ones)
	Rotate(ctx context.Context, uid int64, tokenID int64) (*dto.TokenCreateResponse, error)
	// GetActiveToken gets an active token by ID
	GetActiveToken(ctx context.Context, uid int64, tokenID int64) (*domain.AuthToken, error)
	// RecordAccessLog records a token access log
	RecordAccessLog(ctx context.Context, log *domain.AuthTokenLog) error
	// ListLogs lists access logs for a specific token
	ListLogs(ctx context.Context, uid, tokenID int64, page, pageSize int) ([]*dto.TokenLogResponse, int64, error)
	// UpdateLastUsedAt updates the last used time of a token
	UpdateLastUsedAt(ctx context.Context, tokenID int64) error
	// SetSyncHandler sets the sync hook
	SetSyncHandler(handler func(uid int64, tokenID int64, scope string, kick bool))
}

type tokenService struct {
	tokenRepo    domain.AuthTokenRepository
	logRepo      domain.AuthTokenLogRepository
	tokenManager app.TokenManager
	logger       *zap.Logger
	lastLogMap   sync.Map                                                // TokenID -> time.Time (for 30s rate limiting)
	SyncHandler  func(uid int64, tokenID int64, scope string, kick bool) // Hook for syncing to other modules (like WS)
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
		IssueType:  token.IssueType,
		LastUsedAt: timex.Time(token.LastUsedAt),
		ExpiredAt:  timex.Time(token.ExpiredAt),
		CreatedAt:  timex.Time(token.CreatedAt),
	}
}

func (s *tokenService) Create(ctx context.Context, uid int64, params *dto.TokenIssueRequest) (*dto.TokenCreateResponse, error) {
	// Format scope for 3D-RBAC compatibility
	var formattedScope string
	if params.Protocol != "" || params.Client != "" || params.Function != "" {
		p := params.Protocol
		if p == "" {
			p = "*"
		}
		c := params.Client
		if c == "" {
			c = "*"
		}
		f := params.Function
		if f == "" {
			f = "*"
		}
		formattedScope = fmt.Sprintf("p:%s c:%s f:%s", p, c, f)
	} else {
		// Legacy format (e.g. "p:rest,ws c:ObsidianPlugin f:*")
		formattedScope = "p:" + params.Scope
		if params.ClientType != "" {
			formattedScope += " c:" + params.ClientType
		}
		formattedScope += " f:*"
	}

	t := &domain.AuthToken{
		UID:        uid,
		Scope:      formattedScope,
		ClientType: params.ClientType,
		BoundIP:    params.BoundIP,
		UserAgent:  params.UserAgent,
		Status:     1,
		IssueType:  2, // Manual
		ExpiredAt:  time.Now().Add(time.Duration(params.ExpiredDays) * 24 * time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t, err := s.tokenRepo.Create(ctx, t)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Generate JWT using token_id and a random nonce
	nonce := util.GetRandomString(16)
	tokenStr, err := s.tokenManager.Generate(uid, "", "", t.ID, nonce)
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	// Save nonce to database
	err = s.tokenRepo.UpdateTokenString(ctx, t.ID, nonce)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	t.TokenString = nonce
	res := &dto.TokenCreateResponse{
		TokenResponse: *s.domainToDTO(t),
		TokenString:   tokenStr,
	}
	return res, nil
}

func (s *tokenService) CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error) {
	// Restrict to REST protocol and bind to clientType
	scope := "p:rest c:" + clientType + " f:*"

	t := &domain.AuthToken{
		UID:        uid,
		Scope:      scope,
		ClientType: clientType,
		BoundIP:    ip,
		UserAgent:  userAgent,
		Status:     1,
		IssueType:  1, // Login
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	t, err := s.tokenRepo.Create(ctx, t)
	if err != nil {
		return nil, "", code.ErrorDBQuery.WithDetails(err.Error())
	}

	nonce := util.GetRandomString(16)
	tokenStr, err := s.tokenManager.Generate(uid, "", ip, t.ID, nonce)
	if err != nil {
		return nil, "", code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	err = s.tokenRepo.UpdateTokenString(ctx, t.ID, nonce)
	if err != nil {
		return nil, "", code.ErrorDBQuery.WithDetails(err.Error())
	}

	t.TokenString = nonce

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

func (s *tokenService) Update(ctx context.Context, uid int64, tokenID int64, params *dto.TokenUpdateRequest) error {
	// Need to check if token belongs to user first
	token, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	if token.UID != uid {
		return code.ErrorInvalidAuthToken
	}

	// Only manual tokens can be fully edited
	if token.IssueType != 2 {
		// For login tokens, maybe only allow scope update or just deny
		if params.Scope != "" {
			return s.tokenRepo.UpdateScope(ctx, tokenID, params.Scope)
		}
		return code.ErrorInvalidAuthToken.WithDetails("Only manual tokens can be fully edited")
	}

	// Update fields if provided
	if params.ClientType != "" {
		token.ClientType = params.ClientType
	}
	if params.BoundIP != "" {
		token.BoundIP = params.BoundIP
	}
	if params.UserAgent != "" {
		token.UserAgent = params.UserAgent
	}
	token.ExpiredAt = time.Now().Add(time.Duration(params.ExpiredDays) * 24 * time.Hour)

	// Format scope
	if params.Protocol != "" || params.Client != "" || params.Function != "" {
		p := params.Protocol
		if p == "" {
			p = "*"
		}
		c := params.Client
		if c == "" {
			c = "*"
		}
		f := params.Function
		if f == "" {
			f = "*"
		}
		token.Scope = fmt.Sprintf("p:%s c:%s f:%s", p, c, f)
	} else if params.Scope != "" {
		// If explicit scope is provided, we use it, but check if it's legacy
		if !app.Is3DRBACScope(params.Scope) {
			token.Scope = "p:" + params.Scope + " c:" + token.ClientType + " f:*"
		} else {
			token.Scope = params.Scope
		}
	}

	err = s.tokenRepo.Update(ctx, token)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Trigger sync hook if set
	if s.SyncHandler != nil {
		s.SyncHandler(uid, tokenID, token.Scope, false)
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

	// Trigger sync hook (scope empty means revoked/no permission)
	if s.SyncHandler != nil {
		s.SyncHandler(uid, tokenID, "", true)
	}
	return nil
}

func (s *tokenService) Rotate(ctx context.Context, uid int64, tokenID int64) (*dto.TokenCreateResponse, error) {
	token, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	if token.UID != uid {
		return nil, code.ErrorInvalidAuthToken
	}

	// Only active tokens can be rotated
	if token.Status != 1 {
		return nil, code.ErrorInvalidAuthToken.WithDetails("Token is not active")
	}

	// Prohibit rotation for login tokens (IssueType == 1)
	if token.IssueType == 1 {
		return nil, code.ErrorInvalidAuthToken.WithDetails("Rotation is not allowed for login tokens")
	}

	// Check expiry
	if time.Now().After(token.ExpiredAt) {
		return nil, code.ErrorTokenExpired
	}

	// Generate new JWT with new nonce
	nonce := util.GetRandomString(16)
	tokenStr, err := s.tokenManager.Generate(uid, "", "", token.ID, nonce)
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	// Update nonce in database
	err = s.tokenRepo.UpdateTokenString(ctx, token.ID, nonce)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	token.TokenString = nonce
	res := &dto.TokenCreateResponse{
		TokenResponse: *s.domainToDTO(token),
		TokenString:   tokenStr,
	}

	// Trigger sync hook (just in case scope changed or we need to notify other modules)
	if s.SyncHandler != nil {
		s.SyncHandler(uid, tokenID, token.Scope, true)
	}

	return res, nil
}

func (s *tokenService) RecordAccessLog(ctx context.Context, log *domain.AuthTokenLog) error {
	_ = s.tokenRepo.UpdateLastUsedAt(ctx, log.TokenID)

	// Rate limiting: 30s per TokenID + Protocol
	// 30秒内相同 Token 和协议的连续请求只记录一次
	key := fmt.Sprintf("%d_%s", log.TokenID, log.Protocol)
	if lastTime, ok := s.lastLogMap.Load(key); ok {
		if time.Since(lastTime.(time.Time)) < 30*time.Second {
			return nil
		}
	}
	s.lastLogMap.Store(key, time.Now())

	return s.logRepo.Create(ctx, log)
}

func (s *tokenService) ListLogs(ctx context.Context, uid, tokenID int64, page, pageSize int) ([]*dto.TokenLogResponse, int64, error) {
	// Verify token ownership
	token, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}
	if token.UID != uid {
		return nil, 0, code.ErrorInvalidAuthToken
	}

	logs, count, err := s.logRepo.ListByTokenID(ctx, tokenID, page, pageSize)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var res []*dto.TokenLogResponse
	for _, l := range logs {
		res = append(res, &dto.TokenLogResponse{
			ID:            l.ID,
			Protocol:      l.Protocol,
			Client:        l.Client,
			ClientName:    l.ClientName,
			ClientVersion: l.ClientVersion,
			IP:            l.IP,
			UA:            l.UA,
			StatusCode:    l.StatusCode,
			CreatedAt:     timex.Time(l.CreatedAt),
		})
	}
	return res, count, nil
}

func (s *tokenService) GetActiveToken(ctx context.Context, uid int64, tokenID int64) (*domain.AuthToken, error) {
	token, err := s.tokenRepo.GetByID(ctx, tokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorInvalidUserAuthToken.WithDetails("Token has been revoked or no longer exists")
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	if token.UID != uid {
		return nil, code.ErrorInvalidUserAuthToken.WithDetails("Token does not belong to the authenticated user")
	}
	if token.Status != 1 {
		return nil, code.ErrorInvalidUserAuthToken.WithDetails("Token has been revoked or no longer exists")
	}
	if time.Now().After(token.ExpiredAt) {
		return nil, code.ErrorTokenExpired
	}
	return token, nil
}
func (s *tokenService) UpdateLastUsedAt(ctx context.Context, tokenID int64) error {
	return s.tokenRepo.UpdateLastUsedAt(ctx, tokenID)
}

func (s *tokenService) SetSyncHandler(handler func(uid int64, tokenID int64, scope string, kick bool)) {
	s.SyncHandler = handler
}
