// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
)

// MockNgrokService is a testify/mock implementation of service.NgrokService.
// MockNgrokService 是 service.NgrokService 的 testify/mock 实现。
type MockNgrokService struct {
	mock.Mock
}

// Ensure MockNgrokService implements service.NgrokService at compile time.
// 编译期确保 MockNgrokService 实现了 service.NgrokService 接口。
var _ service.NgrokService = (*MockNgrokService)(nil)

func (m *MockNgrokService) Start(ctx context.Context, addr string) error {
	args := m.Called(ctx, addr)
	return args.Error(0)
}

func (m *MockNgrokService) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockNgrokService) TunnelURL() string {
	args := m.Called()
	return args.String(0)
}
