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

type Note struct {
	ID                      int64      `json:"id" form:"id"`                   // ID
	VaultID                 int64      `json:"vaultId" form:"vaultId"`         // 保险库ID
	Action                  string     `json:"action" form:"action"`           // 操作
	Path                    string     `json:"path" form:"path"`               // 路径
	PathHash                string     `json:"pathHash" form:"pathHash"`       // 路径哈希
	Content                 string     `json:"content" form:"content"`         // 内容
	ContentHash             string     `json:"contentHash" form:"contentHash"` // 内容哈希
	ContentLastSnapshot     string     `json:"contentLastSnapshot" form:"contentLastSnapshot"`
	ContentLastSnapshotHash string     `json:"contentLastSnapshotHash" form:"contentLastSnapshotHash"`
	ClientName              string     `json:"clientName" form:"clientName"`             // 客户端名称
	Version                 int64      `json:"version" form:"version"`                   // 版本
	Size                    int64      `json:"size" form:"size"`                         // 大小
	Ctime                   int64      `json:"ctime" form:"ctime"`                       // 创建时间戳
	Mtime                   int64      `json:"mtime" form:"mtime"`                       // 修改时间戳
	UpdatedTimestamp        int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 更新时间戳
	CreatedAt               timex.Time `json:"createdAt" form:"createdAt"`               // 创建时间
	UpdatedAt               timex.Time `json:"updatedAt" form:"updatedAt"`               // 更新时间

}

type NoteSet struct {
	VaultID     int64  `json:"vaultId" form:"vaultId"`         // 保险库ID
	Action      string `json:"action" form:"action"`           // 操作
	Path        string `json:"path" form:"path"`               // 路径
	PathHash    string `json:"pathHash" form:"pathHash"`       // 路径哈希
	Content     string `json:"content" form:"content"`         // 内容
	ContentHash string `json:"contentHash" form:"contentHash"` // 内容哈希
	ClientName  string `json:"clientName" form:"clientName"`   // 客户端名称
	Size        int64  `json:"size" form:"size"`               // 大小
	Ctime       int64  `json:"ctime" form:"ctime"`             // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`             // 修改时间戳
}

// note 获取笔记查询对象
// 函数名: note
// 函数使用说明: 获取指定用户的笔记表查询对象,内部方法。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *query.Query: 查询对象
func (d *Dao) note(uid int64) *query.Query {
	key := "user_" + strconv.FormatInt(uid, 10)
	return d.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Note")
	}, key+"#note", key)
}

// NoteGetByPathHash 根据路径哈希获取笔记
// 函数名: NoteGetByPathHash
// 函数使用说明: 根据路径哈希值和保险库ID查询笔记记录。
// 参数说明:
//   - hash string: 路径哈希值
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Note: 笔记数据
//   - error: 出错时返回错误
func (d *Dao) NoteGetByPathHash(hash string, vaultID int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(hash),
	).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteGetById 根据ID获取笔记
func (d *Dao) NoteGetById(id int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m, err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}

	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteGetByPath 根据路径获取笔记
// 函数名: NoteGetByPath
// 函数使用说明: 根据笔记路径和保险库ID查询笔记记录。
// 参数说明:
//   - path string: 笔记路径
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Note: 笔记数据
//   - error: 出错时返回错误
func (d *Dao) NoteGetByPath(path string, vaultID int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Eq(path),
	).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteCreate 创建笔记记录
// 函数名: NoteCreate
// 函数使用说明: 在数据库中创建新的笔记记录,自动设置创建时间和更新时间。
// 参数说明:
//   - params *NoteSet: 笔记创建参数
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Note: 创建后的笔记数据
//   - error: 出错时返回错误
func (d *Dao) NoteCreate(params *NoteSet, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m := convert.StructAssign(params, &model.Note{}).(*model.Note)

	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteUpdate 更新笔记记录
// 函数名: NoteUpdate
// 函数使用说明: 根据笔记ID更新笔记记录,自动更新更新时间戳。
// 参数说明:
//   - params *NoteSet: 笔记更新参数
//   - id int64: 笔记ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Note: 更新后的笔记数据
//   - error: 出错时返回错误
func (d *Dao) NoteUpdate(params *NoteSet, id int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m := convert.StructAssign(params, &model.Note{}).(*model.Note)
	//fields := convert.GetCopyStructFields(params, u)
	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.UpdatedAt = timex.Now()
	m.ID = id

	//fields = append(fields, m.UpdatedAt, m.UpdatedTimestamp)

	err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).Select(
		u.ID,
		u.VaultID,
		u.Action,
		u.Path,
		u.PathHash,
		u.Content,
		u.ContentHash,
		u.ClientName,
		u.Size,
		u.Ctime,
		u.Mtime,
		u.UpdatedAt,
		u.UpdatedTimestamp,
	).Save(m)

	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteUpdateMtime 更新笔记的修改时间
// 函数名: NoteUpdateMtime
// 函数使用说明: 只更新笔记的修改时间戳(mtime),同时更新记录的更新时间戳。
// 参数说明:
//   - mtime int64: 新的修改时间戳
//   - id int64: 笔记ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (d *Dao) NoteUpdateMtime(mtime int64, id int64, uid int64) error {
	u := d.note(uid).Note

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.Mtime.Value(mtime),
		u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// NoteUpdateSnapshot 更新笔记的快照内容
func (d *Dao) NoteUpdateSnapshot(snapshot string, snapshotHash string, version int64, id int64, uid int64) error {
	u := d.note(uid).Note
	// 使用 UpdateSimple 更新多个字段
	_, err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).UpdateSimple(
		u.ContentLastSnapshot.Value(snapshot),
		u.ContentLastSnapshotHash.Value(snapshotHash),
		u.Version.Value(version),
	)
	return err
}

// NoteListCount 获取笔记列表数量
// 函数名: NoteListCount
// 函数使用说明: 统计指定保险库中未删除的笔记总数。
// 参数说明:
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//   - keyword string: 搜索关键字
//
// 返回值说明:
//   - int64: 笔记数量
//   - error: 出错时返回错误
func (d *Dao) NoteListCount(vaultID int64, uid int64, keyword string) (int64, error) {
	u := d.note(uid).Note
	q := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	)

	var count int64
	var err error

	if keyword != "" {
		key := "%" + keyword + "%"
		err = q.UnderlyingDB().Where("path LIKE ? OR content LIKE ?", key, key).Count(&count).Error
	} else {
		count, err = q.Order(u.CreatedAt).Count()
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

// NoteList 获取笔记列表
// 函数名: NoteList
// 函数使用说明: 分页查询指定保险库中未删除的笔记列表,按路径和创建时间倒序排列。
// 参数说明:
//   - vaultID int64: 保险库ID
//   - page int: 页码
//   - pageSize int: 每页数量
//   - uid int64: 用户ID
//   - keyword string: 搜索关键字
//
// 返回值说明:
//   - []*Note: 笔记列表
//   - error: 出错时返回错误
func (d *Dao) NoteList(vaultID int64, page int, pageSize int, uid int64, keyword string) ([]*Note, error) {
	u := d.note(uid).Note
	q := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	)

	var modelList []*model.Note
	var err error

	if keyword != "" {
		key := "%" + keyword + "%"
		err = q.UnderlyingDB().Debug().Where("path LIKE ? ", key).
			Order("path DESC, created_at DESC").
			Limit(pageSize).
			Offset(app.GetPageOffset(page, pageSize)).
			Find(&modelList).Error
	} else {
		modelList, err = q.Order(u.Path.Desc(), u.CreatedAt.Desc()).
			Limit(pageSize).
			Offset(app.GetPageOffset(page, pageSize)).
			Find()
	}

	if err != nil {
		return nil, err
	}

	var list []*Note
	for _, m := range modelList {
		list = append(list, convert.StructAssign(m, &Note{}).(*Note))
	}
	return list, nil
}

// NoteListByUpdatedTimestamp 根据更新时间戳获取笔记列表
// 函数名: NoteListByUpdatedTimestamp
// 函数使用说明: 查询指定保险库中更新时间戳大于给定时间戳的所有笔记,按更新时间戳倒序排列。
// 参数说明:
//   - timestamp int64: 时间戳阈值(毫秒)
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - []*Note: 笔记列表
//   - error: 出错时返回错误
func (d *Dao) NoteListByUpdatedTimestamp(timestamp int64, vaultID int64, uid int64) ([]*Note, error) {
	u := d.note(uid).Note
	mList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*Note
	for _, m := range mList {
		list = append(list, convert.StructAssign(m, &Note{}).(*Note))
	}
	return list, nil
}

// NoteListContentUnchanged 获取内容未变更（即未快照）的笔记列表
// 函数名: NoteListContentUnchanged
// 函数使用说明: 查询 content != content_last_snapshot 的笔记。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - []*Note: 笔记列表
//   - error: 出错时返回错误
func (d *Dao) NoteListContentUnchanged(uid int64) ([]*Note, error) {
	u := d.note(uid).Note
	var mList []*model.Note

	// 使用 UnderlyingDB 以支持 raw sql condition "content != content_last_snapshot"
	err := u.WithContext(d.ctx).UnderlyingDB().Where(
		"action != ?", "delete",
	).Where("content != content_last_snapshot").
		Find(&mList).Error

	if err != nil {
		return nil, err
	}

	var list []*Note
	for _, m := range mList {
		list = append(list, convert.StructAssign(m, &Note{}).(*Note))
	}
	return list, nil
}

// NoteListByMtime 根据修改时间戳获取笔记列表
// 函数名: NoteListByMtime
// 函数使用说明: 查询指定保险库中笔记修改时间戳大于给定时间戳的所有笔记,按更新时间戳倒序排列。
// 参数说明:
//   - timestamp int64: 修改时间戳阈值
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - []*Note: 笔记列表
//   - error: 出错时返回错误
func (d *Dao) NoteListByMtime(timestamp int64, vaultID int64, uid int64) ([]*Note, error) {

	u := d.note(uid).Note

	mList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Mtime.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*Note
	for _, m := range mList {
		list = append(list, convert.StructAssign(m, &Note{}).(*Note))
	}
	return list, nil
}

type NoteCountSizeSum struct {
	Size  int64
	Count int64
}

// NoteCountSizeSum 获取笔记数量和大小总和
// 函数名: NoteCountSizeSum
// 函数使用说明: 统计指定保险库中未删除笔记的总数量和总大小。
// 参数说明:
//   - vaultID int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *NoteCountSizeSum: 包含笔记数量和大小总和的结构体
//   - error: 出错时返回错误
func (d *Dao) NoteCountSizeSum(vaultID int64, uid int64) (*NoteCountSizeSum, error) {
	u := d.note(uid).Note

	result := &NoteCountSizeSum{}

	err := u.WithContext(d.ctx).Select(u.Size.Sum().As("size"), u.Size.Count().As("count")).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	).Scan(result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// NoteDeletePhysicalByTime 根据时间物理删除笔记记录
// 函数名: NoteDeletePhysicalByTime
// 函数使用说明: 物理删除更新时间戳小于给定时间戳且已标记为删除的笔记记录。
// 参数说明:
//   - timestamp int64: 时间戳阈值(毫秒)
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (d *Dao) NoteDeletePhysicalByTime(timestamp int64, uid int64) error {
	u := d.note(uid).Note
	_, err := u.WithContext(d.ctx).Where(
		u.Action.Eq("delete"),
		u.UpdatedTimestamp.Lt(timestamp),
	).Delete()
	return err
}

func (d *Dao) NoteDelete(id int64, uid int64) error {

	u := d.note(uid).Note
	_, err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	return nil
}
