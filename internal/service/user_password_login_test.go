package service

import (
	"context"
	"testing"

	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPasswordLoginDisabledRejectsBeforeRepositoryOrTokenAccess(t *testing.T) {
	disabled := false
	repo := new(domainmocks.MockUserRepository) // No calls are expected, including registration checks.
	svc := NewUserService(repo, nil, nil, zap.NewNop(), &ServiceConfig{
		User: UserServiceConfig{PasswordLoginIsEnable: &disabled, RegisterIsEnable: true},
	})
	ctx := context.Background()
	require.False(t, svc.IsRegisterEnabled(ctx))
	for _, credential := range []string{"user", "user@example.invalid", ""} {
		user, err := svc.Login(ctx, &dto.UserLoginRequest{Credentials: credential, Password: "wrong"}, "", "WebGUI", "")
		require.Nil(t, user)
		require.ErrorIs(t, err, code.ErrorUserPasswordLoginIsDisable)
	}
	user, err := svc.Register(ctx, nil, "", "WebGUI", "")
	require.Nil(t, user)
	require.ErrorIs(t, err, code.ErrorUserPasswordLoginIsDisable)
	require.ErrorIs(t, svc.ChangePassword(ctx, 1, nil), code.ErrorUserPasswordLoginIsDisable)
	repo.AssertExpectations(t)
}
