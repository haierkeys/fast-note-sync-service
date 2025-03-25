package dao

import (
	"context"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {

	_, err := global.ConfigLoad("../../config/config-dev.yaml")
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewDBEngine(global.Config.Database)
	if err != nil {
		t.Fatal(err)
	}

	d := New(db, context.Background())

	// 准备测试数据
	params := &NoteSet{
		Vault:    "testVault",
		Action:   "testAction",
		Path:     "testPath",
		PathHash: "testPathHash",
		Content:  "testContent",
		Size:     100,
	}
	uid := int64(1)

	// 调用创建函数
	note, err := d.Create(params, uid)

	dump.P(note)

	// 断言错误为 nil
	assert.Nil(t, err)

	// 断言创建的笔记信息正确
	assert.Equal(t, params.Vault, note.Vault)
	assert.Equal(t, params.Action, note.Action)
	assert.Equal(t, params.Path, note.Path)
	assert.Equal(t, params.PathHash, note.PathHash)
	assert.Equal(t, params.Content, note.Content)
	assert.Equal(t, params.Size, note.Size)
	assert.Equal(t, int64(0), note.IsDeleted)

}

func TestUpdate(t *testing.T) {
	// 创建 Dao 实例
	d := &Dao{}

	// 准备测试数据
	params := &NoteSet{
		Vault:    "updatedVault",
		Action:   "updatedAction",
		Path:     "updatedPath",
		PathHash: "updatedPathHash",
		Content:  "updatedContent",
		Size:     200,
	}
	id := int64(1)
	uid := int64(1)

	// 调用更新函数
	err := d.Update(params, id, uid)

	// 断言错误为 nil
	assert.Nil(t, err)
}

func TestDelete(t *testing.T) {
	// 创建 Dao 实例
	d := &Dao{}

	// 准备测试数据
	id := int64(1)
	uid := int64(1)

	// 调用删除函数
	err := d.Delete(id, uid)

	// 断言错误为 nil
	assert.Nil(t, err)
}
