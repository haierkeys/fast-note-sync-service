// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// newVaultMockRepo returns a fresh MockVaultRepository for each test.
// newVaultMockRepo 为每个测试返回新的 MockVaultRepository。
func newVaultMockRepo() *domainmocks.MockVaultRepository {
	return new(domainmocks.MockVaultRepository)
}

// newVault creates a domain.Vault test fixture.
// newVault 创建 domain.Vault 测试固定数据。
func newVault(id int64, name string) *domain.Vault {
	f := &domainFixture{id: id, name: name}
	return f.toDomain()
}

// vaultDTO checks the DTO contains expected fields.
// vaultDTO 用于验证 DTO 包含预期字段的辅助函数。
func assertVaultDTO(t *testing.T, got *dto.VaultDTO, wantID int64, wantName string) {
	t.Helper()
	assert.NotNil(t, got)
	assert.Equal(t, wantID, got.ID)
	assert.Equal(t, wantName, got.Name)
}

// --- Create ---

// TestVaultService_Create_Success verifies successful vault creation.
// TestVaultService_Create_Success 验证正常创建 Vault 的逻辑。
func TestVaultService_Create_Success(t *testing.T) {
	mockRepo := newVaultMockRepo()

	// GetByName returns not found, so creation proceeds
	// GetByName 返回未找到，允许创建流程继续
	mockRepo.On("GetByName", mock.Anything, "MyVault", int64(1)).
		Return(nil, gorm.ErrRecordNotFound)

	created := newVault(10, "MyVault")
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Vault"), int64(1)).
		Return(created, nil)

	svc := NewVaultService(mockRepo)
	result, err := svc.Create(context.Background(), 1, "MyVault")

	assert.NoError(t, err)
	assertVaultDTO(t, result, 10, "MyVault")
	mockRepo.AssertExpectations(t)
}

// TestVaultService_Create_AlreadyExists verifies error when vault already exists.
// TestVaultService_Create_AlreadyExists 验证 Vault 已存在时返回 ErrorVaultExist。
func TestVaultService_Create_AlreadyExists(t *testing.T) {
	mockRepo := newVaultMockRepo()

	existing := newVault(1, "MyVault")
	mockRepo.On("GetByName", mock.Anything, "MyVault", int64(1)).
		Return(existing, nil)

	svc := NewVaultService(mockRepo)
	_, err := svc.Create(context.Background(), 1, "MyVault")

	assert.ErrorIs(t, err, code.ErrorVaultExist)
	mockRepo.AssertExpectations(t)
}

// --- Get ---

// TestVaultService_Get_Success verifies successful vault retrieval.
// TestVaultService_Get_Success 验证正常获取 Vault 的逻辑。
func TestVaultService_Get_Success(t *testing.T) {
	mockRepo := newVaultMockRepo()

	v := newVault(5, "Notes")
	mockRepo.On("GetByID", mock.Anything, int64(5), int64(1)).
		Return(v, nil)

	svc := NewVaultService(mockRepo)
	result, err := svc.Get(context.Background(), 1, 5)

	assert.NoError(t, err)
	assertVaultDTO(t, result, 5, "Notes")
	mockRepo.AssertExpectations(t)
}

// TestVaultService_Get_NotFound verifies ErrorVaultNotFound when vault is missing.
// TestVaultService_Get_NotFound 验证 Vault 不存在时返回 ErrorVaultNotFound。
func TestVaultService_Get_NotFound(t *testing.T) {
	mockRepo := newVaultMockRepo()

	mockRepo.On("GetByID", mock.Anything, int64(99), int64(1)).
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewVaultService(mockRepo)
	_, err := svc.Get(context.Background(), 1, 99)

	assert.ErrorIs(t, err, code.ErrorVaultNotFound)
	mockRepo.AssertExpectations(t)
}

// --- List ---

// TestVaultService_List_Success verifies successful vault list retrieval.
// TestVaultService_List_Success 验证正常获取 Vault 列表的逻辑。
func TestVaultService_List_Success(t *testing.T) {
	mockRepo := newVaultMockRepo()

	vaults := []*domainFixture{
		{id: 1, name: "Vault-A"},
		{id: 2, name: "Vault-B"},
	}
	mockRepo.On("List", mock.Anything, int64(1)).
		Return(fixturesToDomain(vaults), nil)

	svc := NewVaultService(mockRepo)
	results, err := svc.List(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Vault-A", results[0].Name)
	assert.Equal(t, "Vault-B", results[1].Name)
	mockRepo.AssertExpectations(t)
}

// --- Delete ---

// TestVaultService_Delete_Success verifies successful vault deletion.
// TestVaultService_Delete_Success 验证正常删除 Vault 的逻辑。
func TestVaultService_Delete_Success(t *testing.T) {
	mockRepo := newVaultMockRepo()

	mockRepo.On("Delete", mock.Anything, int64(3), int64(1)).
		Return(nil)

	svc := NewVaultService(mockRepo)
	err := svc.Delete(context.Background(), 1, 3)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- Update ---

// TestVaultService_Update_Success verifies successful vault update.
// TestVaultService_Update_Success 验证正常更新 Vault 的逻辑。
func TestVaultService_Update_Success(t *testing.T) {
	mockRepo := newVaultMockRepo()

	original := newVault(7, "OldName")
	updated := newVault(7, "NewName")

	mockRepo.On("GetByID", mock.Anything, int64(7), int64(1)).
		Return(original, nil).Once()
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Vault"), int64(1)).
		Return(nil)
	// Re-fetch after update
	// 更新后重新获取
	mockRepo.On("GetByID", mock.Anything, int64(7), int64(1)).
		Return(updated, nil).Once()

	svc := NewVaultService(mockRepo)
	result, err := svc.Update(context.Background(), 1, 7, "NewName")

	assert.NoError(t, err)
	assertVaultDTO(t, result, 7, "NewName")
	mockRepo.AssertExpectations(t)
}

// TestVaultService_Update_NotFound verifies error when vault to update does not exist.
// TestVaultService_Update_NotFound 验证要更新的 Vault 不存在时返回错误。
func TestVaultService_Update_NotFound(t *testing.T) {
	mockRepo := newVaultMockRepo()

	mockRepo.On("GetByID", mock.Anything, int64(99), int64(1)).
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewVaultService(mockRepo)
	_, err := svc.Update(context.Background(), 1, 99, "NewName")

	assert.ErrorIs(t, err, code.ErrorVaultNotFound)
	mockRepo.AssertExpectations(t)
}

// --- GetOrCreate ---

// TestVaultService_GetOrCreate_ExistingVault verifies returning existing vault.
// TestVaultService_GetOrCreate_ExistingVault 验证 Vault 已存在时直接返回。
func TestVaultService_GetOrCreate_ExistingVault(t *testing.T) {
	mockRepo := newVaultMockRepo()

	existing := newVault(1, "MyVault")
	mockRepo.On("GetByName", mock.Anything, "MyVault", int64(1)).
		Return(existing, nil)

	svc := NewVaultService(mockRepo)
	result, err := svc.GetOrCreate(context.Background(), 1, "MyVault")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	mockRepo.AssertExpectations(t)
}

// TestVaultService_GetOrCreate_NewVault verifies creating vault when it does not exist.
// TestVaultService_GetOrCreate_NewVault 验证 Vault 不存在时自动创建。
func TestVaultService_GetOrCreate_NewVault(t *testing.T) {
	mockRepo := newVaultMockRepo()

	mockRepo.On("GetByName", mock.Anything, "NewVault", int64(1)).
		Return(nil, gorm.ErrRecordNotFound)

	created := newVault(9, "NewVault")
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Vault"), int64(1)).
		Return(created, nil)

	svc := NewVaultService(mockRepo)
	result, err := svc.GetOrCreate(context.Background(), 1, "NewVault")

	assert.NoError(t, err)
	assert.Equal(t, int64(9), result.ID)
	mockRepo.AssertExpectations(t)
}

// --- UpdateNoteStats / UpdateFileStats ---

// TestVaultService_UpdateNoteStats verifies note stats update delegation.
// TestVaultService_UpdateNoteStats 验证笔记统计更新的委托调用。
func TestVaultService_UpdateNoteStats(t *testing.T) {
	mockRepo := newVaultMockRepo()

	mockRepo.On("UpdateNoteCountSize", mock.Anything, int64(1024), int64(5), int64(1), int64(1)).
		Return(nil)

	svc := NewVaultService(mockRepo)
	err := svc.UpdateNoteStats(context.Background(), 1024, 5, 1, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestVaultService_UpdateFileStats verifies file stats update delegation.
// TestVaultService_UpdateFileStats 验证文件统计更新的委托调用。
func TestVaultService_UpdateFileStats(t *testing.T) {
	mockRepo := newVaultMockRepo()

	mockRepo.On("UpdateFileCountSize", mock.Anything, int64(2048), int64(3), int64(1), int64(1)).
		Return(nil)

	svc := NewVaultService(mockRepo)
	err := svc.UpdateFileStats(context.Background(), 2048, 3, 1, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// --- Test Fixtures ---

// domainFixture is a helper struct for building test domain.Vault instances.
// domainFixture 是用于构建测试 domain.Vault 实例的辅助结构体。
type domainFixture struct {
	id   int64
	name string
}

func (f *domainFixture) toDomain() *domain.Vault {
	return &domain.Vault{
		ID:        f.id,
		Name:      f.name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func fixturesToDomain(fixtures []*domainFixture) []*domain.Vault {
	result := make([]*domain.Vault, len(fixtures))
	for i, f := range fixtures {
		result[i] = f.toDomain()
	}
	return result
}
