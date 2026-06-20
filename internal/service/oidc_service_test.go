package service

import (
	"context"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	internaloidc "github.com/haierkeys/fast-note-sync-service/internal/oidc"
	"gorm.io/gorm"
)

func TestOIDCServiceAuthenticatesExistingIdentity(t *testing.T) {
	userRepo := &fakeOIDCUserRepo{
		byUID: map[int64]*domain.User{
			42: {UID: 42, Email: "oidc@example.com", Username: "oidc-user"},
		},
	}
	identityRepo := &fakeOIDCIdentityRepo{
		byIssuerSubject: map[string]*domain.OIDCIdentity{
			"https://issuer.example|subject-1": {UID: 42, Issuer: "https://issuer.example", Subject: "subject-1"},
		},
	}
	tokenSvc := &fakeOIDCTokenService{}
	svc := NewOIDCService(userRepo, identityRepo, tokenSvc)
	providerConfig := OIDCServiceConfig{
		AutoRegister: false,
		Issuer:       "https://issuer.example",
		UserMapping: OIDCUserMappingConfig{
			SubjectClaim:     "sub",
			EmailClaim:       "email",
			UsernameClaim:    "preferred_username",
			DisplayNameClaim: "name",
		},
	}

	dto, err := svc.Authenticate(context.Background(), providerConfig, internaloidc.Claims{
		Subject:  "subject-1",
		Email:    "oidc@example.com",
		Username: "oidc-user",
	}, "127.0.0.1", "WebGUI", "test-agent")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if dto.UID != 42 || dto.Token != "token-for-42" || dto.TokenID != 100 {
		t.Fatalf("Authenticate() = %#v", dto)
	}
}

func TestOIDCServiceBindsExistingEmailWhenIdentityMissing(t *testing.T) {
	userRepo := &fakeOIDCUserRepo{
		byEmail: map[string]*domain.User{
			"oidc@example.com": {UID: 42, Email: "oidc@example.com", Username: "oidc-user"},
		},
	}
	identityRepo := &fakeOIDCIdentityRepo{byIssuerSubject: map[string]*domain.OIDCIdentity{}}
	svc := NewOIDCService(userRepo, identityRepo, &fakeOIDCTokenService{})
	providerConfig := OIDCServiceConfig{
		Issuer: "https://issuer.example",
		UserMapping: OIDCUserMappingConfig{
			SubjectClaim: "sub",
			EmailClaim:   "email",
		},
	}

	dto, err := svc.Authenticate(context.Background(), providerConfig, internaloidc.Claims{
		Subject: "subject-1",
		Email:   "oidc@example.com",
	}, "127.0.0.1", "WebGUI", "test-agent")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if dto.UID != 42 {
		t.Fatalf("Authenticate().UID = %d, want 42", dto.UID)
	}
	if len(identityRepo.created) != 1 || identityRepo.created[0].UID != 42 {
		t.Fatalf("created identities = %#v", identityRepo.created)
	}
}

func TestOIDCServiceAutoRegistersWhenNoUserMatches(t *testing.T) {
	userRepo := &fakeOIDCUserRepo{byEmail: map[string]*domain.User{}}
	identityRepo := &fakeOIDCIdentityRepo{byIssuerSubject: map[string]*domain.OIDCIdentity{}}
	svc := NewOIDCService(userRepo, identityRepo, &fakeOIDCTokenService{})
	providerConfig := OIDCServiceConfig{
		AutoRegister: true,
		Issuer:       "https://issuer.example",
		UserMapping: OIDCUserMappingConfig{
			SubjectClaim:  "sub",
			EmailClaim:    "email",
			UsernameClaim: "preferred_username",
		},
	}

	dto, err := svc.Authenticate(context.Background(), providerConfig, internaloidc.Claims{
		Subject:  "subject-1",
		Email:    "new@example.com",
		Username: "new-user",
	}, "127.0.0.1", "WebGUI", "test-agent")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if dto.UID != 77 {
		t.Fatalf("Authenticate().UID = %d, want 77", dto.UID)
	}
	if len(userRepo.created) != 1 || userRepo.created[0].Email != "new@example.com" {
		t.Fatalf("created users = %#v", userRepo.created)
	}
}

func TestOIDCServiceAutoRegisterFallsBackToDisplayNameForUsername(t *testing.T) {
	userRepo := &fakeOIDCUserRepo{byEmail: map[string]*domain.User{}}
	identityRepo := &fakeOIDCIdentityRepo{byIssuerSubject: map[string]*domain.OIDCIdentity{}}
	svc := NewOIDCService(userRepo, identityRepo, &fakeOIDCTokenService{})
	providerConfig := OIDCServiceConfig{
		AutoRegister: true,
		Issuer:       "https://issuer.example",
		UserMapping: OIDCUserMappingConfig{
			SubjectClaim:     "sub",
			EmailClaim:       "email",
			UsernameClaim:    "preferred_username",
			DisplayNameClaim: "name",
		},
	}

	_, err := svc.Authenticate(context.Background(), providerConfig, internaloidc.Claims{
		Raw: map[string]interface{}{
			"sub":   "opaque-subject",
			"email": "oidc@example.com",
			"name":  "OIDC User",
		},
	}, "127.0.0.1", "WebGUI", "test-agent")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if len(userRepo.created) != 1 || userRepo.created[0].Username != "OIDC_User" {
		t.Fatalf("created users = %#v, want username OIDC_User", userRepo.created)
	}
}

func TestOIDCServiceRejectsUnknownUserWhenAutoRegisterDisabled(t *testing.T) {
	userRepo := &fakeOIDCUserRepo{byEmail: map[string]*domain.User{}}
	identityRepo := &fakeOIDCIdentityRepo{byIssuerSubject: map[string]*domain.OIDCIdentity{}}
	svc := NewOIDCService(userRepo, identityRepo, &fakeOIDCTokenService{})
	providerConfig := OIDCServiceConfig{
		AutoRegister: false,
		Issuer:       "https://issuer.example",
		UserMapping: OIDCUserMappingConfig{
			SubjectClaim: "sub",
			EmailClaim:   "email",
		},
	}

	if _, err := svc.Authenticate(context.Background(), providerConfig, internaloidc.Claims{
		Subject: "subject-1",
		Email:   "missing@example.com",
	}, "127.0.0.1", "WebGUI", "test-agent"); err == nil {
		t.Fatal("Authenticate() error = nil, want unknown user error")
	}
}

type fakeOIDCUserRepo struct {
	byUID   map[int64]*domain.User
	byEmail map[string]*domain.User
	created []*domain.User
}

func (r *fakeOIDCUserRepo) GetByUID(ctx context.Context, uid int64, onlyActive bool) (*domain.User, error) {
	if user := r.byUID[uid]; user != nil {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeOIDCUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if user := r.byEmail[email]; user != nil {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeOIDCUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeOIDCUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	created := *user
	created.UID = 77
	created.CreatedAt = time.Now()
	created.UpdatedAt = created.CreatedAt
	r.created = append(r.created, &created)
	return &created, nil
}

func (r *fakeOIDCUserRepo) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (r *fakeOIDCUserRepo) UpdatePassword(ctx context.Context, password string, uid int64) error {
	return nil
}

func (r *fakeOIDCUserRepo) GetAll(ctx context.Context) ([]*domain.User, error) {
	return nil, nil
}

func (r *fakeOIDCUserRepo) GetList(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
	return nil, 0, nil
}

func (r *fakeOIDCUserRepo) GetAllUIDs(ctx context.Context) ([]int64, error) {
	return nil, nil
}

type fakeOIDCIdentityRepo struct {
	byIssuerSubject map[string]*domain.OIDCIdentity
	created         []*domain.OIDCIdentity
}

func (r *fakeOIDCIdentityRepo) GetByIssuerSubject(ctx context.Context, issuer, subject string) (*domain.OIDCIdentity, error) {
	if identity := r.byIssuerSubject[issuer+"|"+subject]; identity != nil {
		return identity, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeOIDCIdentityRepo) Create(ctx context.Context, identity *domain.OIDCIdentity) (*domain.OIDCIdentity, error) {
	created := *identity
	created.ID = int64(len(r.created) + 1)
	r.created = append(r.created, &created)
	return &created, nil
}

type fakeOIDCTokenService struct{}

func (s *fakeOIDCTokenService) CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error) {
	return &domain.AuthToken{ID: 100, UID: uid}, "token-for-42", nil
}
