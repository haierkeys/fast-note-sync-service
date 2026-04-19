// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
)

// MockCloudflareService is a testify/mock implementation of service.CloudflareService.
// MockCloudflareService 是 service.CloudflareService 的 testify/mock 实现。
type MockCloudflareService struct {
	mock.Mock
}

// Ensure MockCloudflareService implements service.CloudflareService at compile time.
// 编译期确保 MockCloudflareService 实现了 service.CloudflareService 接口。
var _ service.CloudflareService = (*MockCloudflareService)(nil)

func (m *MockCloudflareService) Start(ctx context.Context, token string, logEnabled bool) error {
	args := m.Called(ctx, token, logEnabled)
	return args.Error(0)
}

func (m *MockCloudflareService) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCloudflareService) TunnelURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCloudflareService) DownloadBinary() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
