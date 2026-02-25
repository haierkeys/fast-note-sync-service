package dao

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

// getContentPath 获取内容存储路径
func (d *Dao) GetNoteFolderPath(uid int64, noteID int64) string {
	return filepath.Join("storage", "vault", fmt.Sprintf("u_%d", uid), "note", fmt.Sprintf("n_%d", noteID))
}

// getSettingFolderPath 获取配置存储路径
func (d *Dao) GetSettingFolderPath(uid int64, settingID int64) string {
	return filepath.Join("storage", "vault", fmt.Sprintf("u_%d", uid), "setting", fmt.Sprintf("s_%d", settingID))
}

// GetFileFolderPath 获取文件目录路径
func (d *Dao) GetFileFolderPath(uid int64, fileID int64) string {
	return filepath.Join("storage", "vault", fmt.Sprintf("u_%d", uid), "file", fmt.Sprintf("f_%d", fileID))
}

// GetNoteHistoryFolderPath 获取笔记历史存储路径
func (d *Dao) GetNoteHistoryFolderPath(uid int64, historyID int64) string {
	return filepath.Join("storage", "vault", fmt.Sprintf("u_%d", uid), "history", fmt.Sprintf("h_%d", historyID))
}

// saveContentToFile 保存内容到文件
func (d *Dao) SaveContentToFile(folderPath string, fileName string, content string) error {
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(folderPath, fileName)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// loadContentFromFile 从文件加载内容
// 返回值: 内容, 是否存在, 错误
func (d *Dao) LoadContentFromFile(folderPath string, fileName string) (string, bool, error) {
	filePath := filepath.Join(folderPath, fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return string(content), true, nil
}

// removeContentFolder 删除内容文件夹
func (d *Dao) RemoveContentFolder(folderPath string) error {
	if fileurl.IsExist(folderPath) {
		return os.RemoveAll(folderPath)
	}
	return nil
}
