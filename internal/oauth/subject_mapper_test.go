package oauth

import (
	"context"
	"errors"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
)

type fakeUserByEmailRepository struct {
	users map[string]*domain.User
	calls []string
	err   error
}

func (r *fakeUserByEmailRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.calls = append(r.calls, email)
	if r.err != nil {
		return nil, r.err
	}
	return r.users[email], nil
}

func TestSubjectMapper_emailModeUsesConfiguredClaim(t *testing.T) {
	repo := &fakeUserByEmailRepository{
		users: map[string]*domain.User{
			"oauth@example.com": {UID: 42, Email: "oauth@example.com"},
		},
	}
	mapper := NewSubjectMapper(repo, SubjectMapperConfig{
		Mode:       SubjectMapperModeEmail,
		EmailClaim: "preferred_email",
	})

	uid, err := mapper.Map(context.Background(), map[string]interface{}{
		"preferred_email": "oauth@example.com",
	})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	if uid != 42 {
		t.Fatalf("Map() uid = %d, want 42", uid)
	}
	if len(repo.calls) != 1 || repo.calls[0] != "oauth@example.com" {
		t.Fatalf("GetByEmail calls = %#v, want oauth@example.com", repo.calls)
	}
}

func TestSubjectMapper_emailModeMissingUserReturnsSubjectNotFound(t *testing.T) {
	mapper := NewSubjectMapper(&fakeUserByEmailRepository{}, SubjectMapperConfig{
		Mode: SubjectMapperModeEmail,
	})

	_, err := mapper.Map(context.Background(), map[string]interface{}{
		"email": "missing@example.com",
	})
	if !errors.Is(err, ErrSubjectNotFound) {
		t.Fatalf("Map() error = %v, want ErrSubjectNotFound", err)
	}
}

func TestSubjectMapper_fixedUIDModeUsesConfiguredUID(t *testing.T) {
	mapper := NewSubjectMapper(&fakeUserByEmailRepository{}, SubjectMapperConfig{
		Mode:     SubjectMapperModeFixedUID,
		FixedUID: 7,
	})

	uid, err := mapper.Map(context.Background(), nil)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	if uid != 7 {
		t.Fatalf("Map() uid = %d, want 7", uid)
	}
}

func TestSubjectMapper_emailOrFixedUIDFallsBackWhenEmailMissing(t *testing.T) {
	mapper := NewSubjectMapper(&fakeUserByEmailRepository{}, SubjectMapperConfig{
		Mode:     SubjectMapperModeEmailOrFixedUID,
		FixedUID: 8,
	})

	uid, err := mapper.Map(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	if uid != 8 {
		t.Fatalf("Map() uid = %d, want 8", uid)
	}
}
