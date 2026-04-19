// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
)

// MockNoteLinkService is a testify/mock implementation of service.NoteLinkService.
// MockNoteLinkService 是 service.NoteLinkService 的 testify/mock 实现。
type MockNoteLinkService struct {
	mock.Mock
}

// Ensure MockNoteLinkService implements service.NoteLinkService at compile time.
// 编译期确保 MockNoteLinkService 实现了 service.NoteLinkService 接口。
var _ service.NoteLinkService = (*MockNoteLinkService)(nil)

func (m *MockNoteLinkService) GetBacklinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteLinkItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteLinkService) GetOutlinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteLinkItem), args.Error(1)
	}
	return nil, args.Error(1)
}
