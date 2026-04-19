// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
)

// MockConflictService is a testify/mock implementation of service.ConflictService.
// MockConflictService 是 service.ConflictService 的 testify/mock 实现。
type MockConflictService struct {
	mock.Mock
}

// Ensure MockConflictService implements service.ConflictService at compile time.
// 编译期确保 MockConflictService 实现了 service.ConflictService 接口。
var _ service.ConflictService = (*MockConflictService)(nil)

func (m *MockConflictService) CreateConflictFile(ctx context.Context, uid int64, params *dto.ConflictFileRequest) (*dto.ConflictFileResponse, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.ConflictFileResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
