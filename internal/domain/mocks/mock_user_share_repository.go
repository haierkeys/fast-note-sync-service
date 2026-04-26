// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockUserShareRepository is a testify mock for domain.UserShareRepository.
// MockUserShareRepository 是 domain.UserShareRepository 的 testify mock 实现。
type MockUserShareRepository struct {
	mock.Mock
}

func (m *MockUserShareRepository) Create(ctx context.Context, uid int64, share *domain.UserShare) error {
	args := m.Called(ctx, uid, share)
	return args.Error(0)
}

func (m *MockUserShareRepository) GetByID(ctx context.Context, uid int64, id int64) (*domain.UserShare, error) {
	args := m.Called(ctx, uid, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserShare), args.Error(1)
}

func (m *MockUserShareRepository) GetByPath(ctx context.Context, uid int64, vaultID int64, pathHash string) (*domain.UserShare, error) {
	args := m.Called(ctx, uid, vaultID, pathHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserShare), args.Error(1)
}

func (m *MockUserShareRepository) GetByRes(ctx context.Context, uid int64, resType string, resID int64) (*domain.UserShare, error) {
	args := m.Called(ctx, uid, resType, resID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserShare), args.Error(1)
}

func (m *MockUserShareRepository) UpdateResources(ctx context.Context, uid int64, id int64, resources map[string][]string) error {
	args := m.Called(ctx, uid, id, resources)
	return args.Error(0)
}

func (m *MockUserShareRepository) UpdateStatus(ctx context.Context, uid int64, id int64, status int64) error {
	args := m.Called(ctx, uid, id, status)
	return args.Error(0)
}

func (m *MockUserShareRepository) UpdateStatusByRes(ctx context.Context, uid int64, resType string, resID int64, status int64) error {
	args := m.Called(ctx, uid, resType, resID, status)
	return args.Error(0)
}

func (m *MockUserShareRepository) UpdateViewStats(ctx context.Context, uid int64, id int64, viewCountIncr int64, lastViewedAt time.Time) error {
	args := m.Called(ctx, uid, id, viewCountIncr, lastViewedAt)
	return args.Error(0)
}

func (m *MockUserShareRepository) UpdatePassword(ctx context.Context, uid int64, id int64, password string) error {
	args := m.Called(ctx, uid, id, password)
	return args.Error(0)
}

func (m *MockUserShareRepository) UpdateShortLink(ctx context.Context, uid int64, id int64, shortLink string) error {
	args := m.Called(ctx, uid, id, shortLink)
	return args.Error(0)
}

func (m *MockUserShareRepository) ListByUID(ctx context.Context, uid int64, sortBy string, sortOrder string, offset, limit int) ([]*domain.UserShare, error) {
	args := m.Called(ctx, uid, sortBy, sortOrder, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserShare), args.Error(1)
}

func (m *MockUserShareRepository) CountByUID(ctx context.Context, uid int64) (int64, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserShareRepository) ListActiveNoteResIDs(ctx context.Context, uid int64) ([]int64, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockUserShareRepository) ListChangedNoteResIDs(ctx context.Context, uid int64, since time.Time) (active []int64, revoked []int64, err error) {
	args := m.Called(ctx, uid, since)
	var a, r []int64
	if args.Get(0) != nil {
		a = args.Get(0).([]int64)
	}
	if args.Get(1) != nil {
		r = args.Get(1).([]int64)
	}
	return a, r, args.Error(2)
}

func (m *MockUserShareRepository) MigrateResID(ctx context.Context, uid int64, oldResID int64, newResID int64) error {
	args := m.Called(ctx, uid, oldResID, newResID)
	return args.Error(0)
}

func (m *MockUserShareRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

// Compile-time check: MockUserShareRepository must implement domain.UserShareRepository.
// 编译时检查：MockUserShareRepository 必须实现 domain.UserShareRepository 接口。
var _ domain.UserShareRepository = (*MockUserShareRepository)(nil)
