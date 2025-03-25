package service

import (
	"github.com/haierkeys/obsidian-better-sync-service/internal/dao"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/timex"
)

type Note struct {
	ID        int64      `json:"id" form:"id"`               // 主键ID
	Vault     string     `json:"vault" form:"vault"`         // 保险库名称或标识
	Action    string     `json:"action" form:"action"`       // 操作类型
	Path      string     `json:"path" form:"path"`           // 路径信息
	PathHash  string     `json:"pathHash" form:"pathHash"`   // 路径哈希值
	Content   string     `json:"content" form:"content"`     // 内容详情
	Size      int64      `json:"size" form:"size"`           // 内容大小，不能为空
	CreatedAt timex.Time `json:"createdAt" form:"createdAt"` // 创建时间，自动填充当前时间
	UpdatedAt timex.Time `json:"updatedAt" form:"updatedAt"` // 更新时间，自动填充当前时间
}

/**
* FileCreateRequestParams
* @Description        文件创建请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type FileCreateRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
	Content  string `json:"content" form:"content" binding:"required"`
}

/**
* FileModifyRequestParams
* @Description        文件修改请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type FileModifyRequestParams struct {
	Vault    string `json:"vault" form:"vault"  binding:"required"`
	Path     string `json:"path" form:"path"  binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"  binding:"required"`
	Content  string `json:"content" form:"content"  binding:"required"`
}

/**
* ContentModifyRequestParams
* @Description        文件内容修改请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type ContentModifyRequestParams struct {
	Vault    string `json:"vault" form:"vault"  binding:"required"`
	Path     string `json:"path" form:"path"  binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"  binding:"required"`
	Content  string `json:"content" form:"content"  binding:"required"`
}

/**
* FileDeleteRequestParams
* @Description        文件删除请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type FileDeleteRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
}

/**
* FileCreate
* @Description        创建文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *FileCreateRequestParams  文件创建请求参数
* @Return             int64  文件ID
* @Return             error  错误信息
 */
func (svc *Service) FileCreate(uid int64, params *FileCreateRequestParams) (int64, error) {
	note := &dao.NoteSet{
		Vault:    params.Vault,
		Action:   "create",
		Path:     params.Path,
		PathHash: params.PathHash,
		Content:  params.Content,
		Size:     int64(len(params.Content)),
	}
	user, err := svc.dao.NoteCreate(note, uid)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

/**
* FileModify
* @Description        修改文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *FileModifyRequestParams  文件修改请求参数
* @Return             error  错误信息
 */
func (svc *Service) FileModify(uid int64, params *FileModifyRequestParams) error {
	node, err := svc.dao.NoteGetByPathHash(params.PathHash, params.Vault, uid)
	if err != nil {
		return err
	}
	note := &dao.NoteSet{
		Vault:    node.Vault,
		Action:   "modify",
		Path:     node.Path,
		PathHash: node.PathHash,
		Size:     int64(len(params.Content)),
	}
	err = svc.dao.NoteUpdate(note, node.ID, uid)
	if err != nil {
		return err
	}
	return nil
}

/**
* ContentModify
* @Description        修改文件内容
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *ContentModifyRequestParams  文件内容修改请求参数
* @Return             error  错误信息
 */
func (svc *Service) ContentModify(uid int64, params *ContentModifyRequestParams) error {
	node, err := svc.dao.NoteGetByPathHash(params.PathHash, params.Vault, uid)
	if err != nil {
		return err
	}
	note := &dao.NoteSet{
		Vault:    node.Vault,
		Action:   "modify",
		Path:     node.Path,
		PathHash: node.PathHash,
		Size:     int64(len(params.Content)),
	}
	err = svc.dao.NoteUpdate(note, node.ID, uid)
	if err != nil {
		return err
	}
	return nil
}

/**
* FileDelete
* @Description        删除文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *FileDeleteRequestParams  文件删除请求参数
* @Return             error  错误信息
 */
func (svc *Service) FileDelete(uid int64, params *FileDeleteRequestParams) error {
	node, err := svc.dao.NoteGetByPathHash(params.PathHash, params.Vault, uid)
	if err != nil {
		return err
	}
	note := &dao.NoteSet{
		Vault:    node.Vault,
		Action:   "delete",
		Path:     node.Path,
		PathHash: node.PathHash,
		Content:  "",
		Size:     0,
	}
	err = svc.dao.NoteUpdate(note, node.ID, uid)
	if err != nil {
		return err
	}
	return nil
}
