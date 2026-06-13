package oauth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
)

const (
	SubjectMapperModeEmail           = "email"
	SubjectMapperModeFixedUID        = "fixed_uid"
	SubjectMapperModeEmailOrFixedUID = "email_or_fixed_uid"
)

type UserByEmailRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type SubjectMapperConfig struct {
	Mode       string
	EmailClaim string
	FixedUID   int64
}

type SubjectMapper struct {
	repo   UserByEmailRepository
	config SubjectMapperConfig
}

func NewSubjectMapper(repo UserByEmailRepository, config SubjectMapperConfig) *SubjectMapper {
	return &SubjectMapper{
		repo:   repo,
		config: config,
	}
}

func (m *SubjectMapper) Map(ctx context.Context, claims map[string]interface{}) (int64, error) {
	mode := strings.TrimSpace(m.config.Mode)
	if mode == "" {
		mode = SubjectMapperModeEmail
	}

	switch mode {
	case SubjectMapperModeEmail:
		return m.mapEmail(ctx, claims)
	case SubjectMapperModeFixedUID:
		return m.fixedUID()
	case SubjectMapperModeEmailOrFixedUID:
		uid, err := m.mapEmail(ctx, claims)
		if err == nil {
			return uid, nil
		}
		if !isSubjectFallbackError(err) {
			return 0, err
		}
		return m.fixedUID()
	default:
		return 0, fmt.Errorf("%w: unsupported subject mapper mode %q", ErrConfig, mode)
	}
}

func (m *SubjectMapper) mapEmail(ctx context.Context, claims map[string]interface{}) (int64, error) {
	if m.repo == nil {
		return 0, fmt.Errorf("%w: user repository is required", ErrConfig)
	}

	claimName := strings.TrimSpace(m.config.EmailClaim)
	if claimName == "" {
		claimName = "email"
	}

	email, _ := claims[claimName].(string)
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return 0, fmt.Errorf("%w: claim %q is empty", ErrSubjectNotFound, claimName)
	}

	user, err := m.repo.GetByEmail(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrSubjectNotFound, err)
	}
	if user == nil || user.UID == 0 {
		return 0, fmt.Errorf("%w: email %q", ErrSubjectNotFound, email)
	}

	return user.UID, nil
}

func (m *SubjectMapper) fixedUID() (int64, error) {
	if m.config.FixedUID <= 0 {
		return 0, fmt.Errorf("%w: fixed uid is required", ErrConfig)
	}
	return m.config.FixedUID, nil
}

func isSubjectFallbackError(err error) bool {
	return errors.Is(err, ErrSubjectNotFound)
}
