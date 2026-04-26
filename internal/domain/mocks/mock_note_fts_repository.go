package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNoteFTSRepository is a testify mock for domain.NoteFTSRepository.
type MockNoteFTSRepository struct {
	mock.Mock
}

func (m *MockNoteFTSRepository) Upsert(ctx context.Context, noteID int64, path, content string, uid int64) error {
	args := m.Called(ctx, noteID, path, content, uid)
	return args.Error(0)
}

func (m *MockNoteFTSRepository) Delete(ctx context.Context, noteID int64, uid int64) error {
	args := m.Called(ctx, noteID, uid)
	return args.Error(0)
}

func (m *MockNoteFTSRepository) Search(ctx context.Context, keyword string, vaultID, uid int64, limit, offset int) ([]int64, error) {
	args := m.Called(ctx, keyword, vaultID, uid, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockNoteFTSRepository) SearchCount(ctx context.Context, keyword string, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, keyword, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNoteFTSRepository) RebuildIndex(ctx context.Context, uid int64) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *MockNoteFTSRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

var _ domain.NoteFTSRepository = (*MockNoteFTSRepository)(nil)
