package dao

import (
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type File struct {
	ID               int64      `json:"id" form:"id"`                             // ID
	VaultID          int64      `json:"vaultId" form:"vaultId"`                   // 保险库ID
	Action           string     `json:"action" form:"action"`                     // 操作
	Path             string     `json:"path" form:"path"`                         // 路径
	PathHash         string     `json:"pathHash" form:"pathHash"`                 // 路径哈希
	ContentHash      string     `json:"contentHash" form:"contentHash"`           // 内容哈希
	SavePath         string     `json:"savePath" form:"savePath"`                 // 保存路径
	Size             int64      `json:"size" form:"size"`                         // 大小
	Ctime            int64      `json:"ctime" form:"ctime"`                       // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`                       // 修改时间戳
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 更新时间戳
	CreatedAt        timex.Time `json:"createdAt" form:"createdAt"`               // 创建时间
	UpdatedAt        timex.Time `json:"updatedAt" form:"updatedAt"`               // 更新时间
}

type FileSet struct {
	VaultID     int64  `json:"vaultId" form:"vaultId"`         // 保险库ID
	Action      string `json:"action" form:"action"`           // 操作
	Path        string `json:"path" form:"path"`               // 路径
	PathHash    string `json:"pathHash" form:"pathHash"`       // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"` // 内容哈希
	SavePath    string `json:"savePath" form:"savePath"`       // 保存路径
	Size        int64  `json:"size" form:"size"`               // 大小
	Ctime       int64  `json:"ctime" form:"ctime"`             // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`             // 修改时间戳
}

// file 获取文件查询对象
// 函数名: file
// 函数使用说明: 获取指定用户的文件表查询对象,内部方法。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *query.Query: 查询对象
func (d *Dao) file(uid int64) *query.Query {
	key := "user_" + strconv.FormatInt(uid, 10)
	return d.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "File")
	}, key+"#file", key)
}

// FileGetByPathHash 根据路径哈希获取文件
// 函数名: FileGetByPathHash
// 函数使用说明: 根据路径哈希值和保险库ID查询文件记录。
// 参数说明:
//   - hash string: 路径哈希值
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *File: 文件数据
//   - error: 出错时返回错误
func (d *Dao) FileGetByPathHash(hash string, vaultID int64, uid int64) (*File, error) {
	u := d.file(uid).File
	m, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(hash),
	).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &File{}).(*File), nil
}

// FileGetByPath 根据路径获取文件
// 函数名: FileGetByPath
// 函数使用说明: 根据文件路径和保险库ID查询文件记录。
// 参数说明:
//   - path string: 文件路径
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *File: 文件数据
//   - error: 出错时返回错误
func (d *Dao) FileGetByPath(path string, vaultID int64, uid int64) (*File, error) {
	u := d.file(uid).File
	m, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Eq(path),
	).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &File{}).(*File), nil
}

// FileCreate 创建文件记录
// 函数名: FileCreate
// 函数使用说明: 在数据库中创建新的文件记录,自动设置创建时间和更新时间。
// 参数说明:
//   - params *FileSet: 文件创建参数
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *File: 创建后的文件数据
//   - error: 出错时返回错误
func (d *Dao) FileCreate(params *FileSet, uid int64) (*File, error) {
	u := d.file(uid).File
	m := convert.StructAssign(params, &model.File{}).(*model.File)

	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &File{}).(*File), nil
}

// FileUpdate 更新文件记录
// 函数名: FileUpdate
// 函数使用说明: 根据文件ID更新文件记录,自动更新更新时间戳。
// 参数说明:
//   - params *FileSet: 文件更新参数
//   - id int64: 文件ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *File: 更新后的文件数据
//   - error: 出错时返回错误
func (d *Dao) FileUpdate(params *FileSet, id int64, uid int64) (*File, error) {
	u := d.file(uid).File
	m := convert.StructAssign(params, &model.File{}).(*model.File)
	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.UpdatedAt = timex.Now()
	m.ID = id
	err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).Save(m)

	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &File{}).(*File), nil
}

// FileUpdateMtime 更新文件的修改时间
// 函数名: FileUpdateMtime
// 函数使用说明: 只更新文件的修改时间戳(mtime),同时更新记录的更新时间戳。
// 参数说明:
//   - mtime int64: 新的修改时间戳
//   - id int64: 文件ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (d *Dao) FileUpdateMtime(mtime int64, id int64, uid int64) error {
	u := d.file(uid).File

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.Mtime.Value(mtime),
		u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// FileListCount 获取文件列表数量
// 函数名: FileListCount
// 函数使用说明: 统计指定保险库中未删除的文件总数。
// 参数说明:
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - int64: 文件数量
//   - error: 出错时返回错误
func (d *Dao) FileListCount(vaultID int64, uid int64) (int64, error) {
	u := d.file(uid).File
	count, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	).Order(u.CreatedAt).
		Count()

	if err != nil {
		return 0, err
	}

	return count, nil
}

// FileList 获取文件列表
// 函数名: FileList
// 函数使用说明: 分页查询指定保险库中未删除的文件列表,按路径和创建时间倒序排列。
// 参数说明:
//   - vaultID int64: 保险库ID
//   - page int: 页码
//   - pageSize int: 每页数量
//   - uid int64: 用户ID
//
// 返回值说明:
//   - []*File: 文件列表
//   - error: 出错时返回错误
func (d *Dao) FileList(vaultID int64, page int, pageSize int, uid int64) ([]*File, error) {
	u := d.file(uid).File
	modelList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	).Order(u.Path.Desc(), u.CreatedAt.Desc()).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*File
	for _, m := range modelList {
		list = append(list, convert.StructAssign(m, &File{}).(*File))
	}
	return list, nil
}

// FileListByUpdatedTimestamp 根据更新时间戳获取文件列表
// 函数名: FileListByUpdatedTimestamp
// 函数使用说明: 查询指定保险库中更新时间戳大于给定时间戳的所有文件,按更新时间戳倒序排列。
// 参数说明:
//   - timestamp int64: 时间戳阈值(毫秒)
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - []*File: 文件列表
//   - error: 出错时返回错误
func (d *Dao) FileListByUpdatedTimestamp(timestamp int64, vaultID int64, uid int64) ([]*File, error) {
	u := d.file(uid).File
	mList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*File
	for _, m := range mList {
		list = append(list, convert.StructAssign(m, &File{}).(*File))
	}
	return list, nil
}

// FileListByMtime 根据修改时间戳获取文件列表
// 函数名: FileListByMtime
// 函数使用说明: 查询指定保险库中文件修改时间戳大于给定时间戳的所有文件,按更新时间戳倒序排列。
// 参数说明:
//   - timestamp int64: 修改时间戳阈值
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - []*File: 文件列表
//   - error: 出错时返回错误
func (d *Dao) FileListByMtime(timestamp int64, vaultID int64, uid int64) ([]*File, error) {

	u := d.file(uid).File

	mList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Mtime.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*File
	for _, m := range mList {
		list = append(list, convert.StructAssign(m, &File{}).(*File))
	}
	return list, nil
}

type FileCountSizeSum struct {
	Size  int64
	Count int64
}

// FileCountSizeSum 获取文件数量和大小总和
// 函数名: FileCountSizeSum
// 函数使用说明: 统计指定保险库中未删除文件的总数量和总大小。
// 参数说明:
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *FileCountSizeSum: 包含文件数量和大小总和的结构体
//   - error: 出错时返回错误
func (d *Dao) FileCountSizeSum(vaultID int64, uid int64) (*FileCountSizeSum, error) {
	u := d.file(uid).File

	result := &FileCountSizeSum{}

	err := u.WithContext(d.ctx).Select(u.Size.Sum().As("size"), u.Size.Count().As("count")).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	).Scan(result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// FileDeletePhysicalByTime 根据时间物理删除文件记录
// 函数名: FileDeletePhysicalByTime
// 函数使用说明: 物理删除更新时间戳小于给定时间戳且已标记为删除的文件记录。
// 参数说明:
//   - timestamp int64: 时间戳阈值(毫秒)
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (d *Dao) FileDeletePhysicalByTime(timestamp int64, uid int64) error {
	u := d.file(uid).File

	_, err := u.WithContext(d.ctx).Where(
		u.Action.Eq("delete"),
		u.UpdatedTimestamp.Lt(timestamp),
	).Delete()
	return err
}
