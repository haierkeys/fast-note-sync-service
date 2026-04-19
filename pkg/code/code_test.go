package code

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCode_Immutability_And_Methods(t *testing.T) {
	// 临时注册一个错误码用于测试，避免和 common 里的冲突
	codeVal := 999901
	msgZh := "测试错误"
	msgEn := "Test Error"
	
	zh_cn_messages[codeVal] = msgZh
	en_messages[codeVal] = msgEn

	c := NewError(codeVal)
	assert.Equal(t, codeVal, c.Code())
	assert.False(t, c.Status())
	// Default language logic depends on getLang config, let's just make sure Msg() won't panic
	assert.NotEmpty(t, c.Msg())
	
	// Test Immutability
	cWithData := c.WithData(map[string]string{"foo": "bar"})
	assert.NotEqual(t, fmt.Sprintf("%p", c), fmt.Sprintf("%p", cWithData))
	assert.False(t, c.HaveData())
	assert.True(t, cWithData.HaveData())
	assert.Equal(t, map[string]string{"foo": "bar"}, cWithData.Data())

	cWithVault := c.WithVault("test-vault")
	assert.NotEqual(t, fmt.Sprintf("%p", c), fmt.Sprintf("%p", cWithVault))
	assert.False(t, c.HaveVault())
	assert.True(t, cWithVault.HaveVault())
	assert.Equal(t, "test-vault", cWithVault.Vault())

	cWithDetails := c.WithDetails("detail1", "detail2")
	assert.NotEqual(t, fmt.Sprintf("%p", c), fmt.Sprintf("%p", cWithDetails))
	assert.False(t, c.HaveDetails())
	assert.True(t, cWithDetails.HaveDetails())
	assert.Equal(t, []string{"detail1", "detail2"}, cWithDetails.Details())

	cWithContext := c.WithContext("test-context")
	assert.NotEqual(t, fmt.Sprintf("%p", c), fmt.Sprintf("%p", cWithContext))
	assert.False(t, c.HaveContext())
	assert.True(t, cWithContext.HaveContext())
	assert.Equal(t, "test-context", cWithContext.Context())
}

func TestNewError_DuplicatePanic(t *testing.T) {
	// 期望注册重复的错误码时触发 panic
	codeVal := 999902
	zh_cn_messages[codeVal] = "测试崩溃"
	en_messages[codeVal] = "Test Panic"

	NewError(codeVal) // 第一次注册成功
	assert.Panics(t, func() {
		NewError(codeVal) // 第二次注册失败并 Panic
	})
}

func TestNewSuss(t *testing.T) {
	codeVal := 999903
	zh_cn_messages[codeVal] = "测试成功"
	en_messages[codeVal] = "Test Success"

	s := NewSuss(codeVal)
	assert.Equal(t, codeVal, s.Code())
	assert.True(t, s.Status())
	assert.NotEmpty(t, s.Msg())

	assert.Panics(t, func() {
		NewSuss(codeVal) // duplicate suss code
	})
}
