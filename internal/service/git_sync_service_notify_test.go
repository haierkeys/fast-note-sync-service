package service

import (
	"context"
	"testing"
	"time"

	appconfig "github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newTestGitSyncService(repo domain.GitSyncRepository) *gitSyncService {
	vaultRepo := new(domainmocks.MockVaultRepository)
	vaultRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.Vault{Name: "test-vault"}, nil)

	svc := NewGitSyncService(
		repo,
		new(domainmocks.MockNoteRepository),
		new(domainmocks.MockFolderRepository),
		new(domainmocks.MockFileRepository),
		vaultRepo,
		new(domainmocks.MockSettingRepository),
		&appconfig.GitConfig{},
		zap.NewNop(),
	)
	return svc.(*gitSyncService)
}

// TestGitSyncService_NotifyUpdated_CachesNoConfigResult verifies the P2 fix: once
// NotifyUpdated has determined a vault has no enabled git sync config, subsequent calls
// for the same (uid, vaultID) within the TTL skip the ListByVaultID DB query entirely.
func TestGitSyncService_NotifyUpdated_CachesNoConfigResult(t *testing.T) {
	repo := new(domainmocks.MockGitSyncRepository)
	repo.On("ListByVaultID", mock.Anything, int64(1), int64(1)).
		Return([]*domain.GitSyncConfig{}, nil).Once()

	svc := newTestGitSyncService(repo)

	svc.NotifyUpdated(1, 1)
	svc.NotifyUpdated(1, 1)
	svc.NotifyUpdated(1, 1)

	repo.AssertNumberOfCalls(t, "ListByVaultID", 1)
}

// TestGitSyncService_NotifyUpdated_NeverCachesEnabledResult verifies that when a vault
// does have an enabled config, the cache never short-circuits the query (since real
// per-call config data, e.g. Delay, is needed to schedule the debounce timer correctly).
func TestGitSyncService_NotifyUpdated_NeverCachesEnabledResult(t *testing.T) {
	repo := new(domainmocks.MockGitSyncRepository)
	repo.On("ListByVaultID", mock.Anything, int64(2), int64(1)).
		Return([]*domain.GitSyncConfig{
			{ID: 10, UID: 1, VaultID: 2, IsEnabled: true, Delay: 30},
		}, nil)

	svc := newTestGitSyncService(repo)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = svc.Shutdown(ctx)
	}()

	svc.NotifyUpdated(1, 2)
	svc.NotifyUpdated(1, 2)

	repo.AssertNumberOfCalls(t, "ListByVaultID", 2)
}

// TestGitSyncService_UpdateConfig_InvalidatesCache verifies that UpdateConfig invalidates
// the NotifyUpdated cache, so a vault that was cached as "no enabled config" is re-checked
// immediately after its config is enabled, instead of waiting out the TTL.
func TestGitSyncService_UpdateConfig_InvalidatesCache(t *testing.T) {
	repo := new(domainmocks.MockGitSyncRepository)

	// First NotifyUpdated call: no configs yet, gets cached as "disabled".
	repo.On("ListByVaultID", mock.Anything, int64(3), int64(1)).
		Return([]*domain.GitSyncConfig{}, nil).Once()

	svc := newTestGitSyncService(repo)
	svc.NotifyUpdated(1, 3)

	// Now a config is created/enabled for this vault via UpdateConfig.
	saved := &domain.GitSyncConfig{ID: 20, UID: 1, VaultID: 3, IsEnabled: true, Delay: 5}
	repo.On("Save", mock.Anything, mock.Anything, int64(1)).Return(saved, nil)

	_, err := svc.UpdateConfig(context.Background(), 1, &dto.GitSyncConfigRequest{
		RepoURL:   "https://example.com/repo.git",
		IsEnabled: true,
		Delay:     5,
	})
	assert.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = svc.Shutdown(ctx)
	}()

	// Second NotifyUpdated call: cache must have been invalidated by UpdateConfig, so this
	// must re-query the DB (which now returns the enabled config) instead of trusting the
	// stale "disabled" cache entry.
	repo.On("ListByVaultID", mock.Anything, int64(3), int64(1)).
		Return([]*domain.GitSyncConfig{saved}, nil).Once()

	svc.NotifyUpdated(1, 3)

	repo.AssertExpectations(t)
}
