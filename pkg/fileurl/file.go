package fileurl

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileType int

const ImageType FileType = iota + 1

// IsFile determines if the given path is a file
// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// IsDir determines if the given path is a directory
// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()

}

// GetFileName gets file path
// GetFileName 获取文件路径
func GetFileName(name string) string {
	ext := GetFileExt(name)
	fileName := strings.TrimSuffix(name, ext)
	// fileName = util.EncodeMD5(fileName)
	return fileName + ext
}

func GetFileNameOrRandom(fileName string) string {
	// Attachments uploaded via clipboard are given a default name
	// 通过剪切板上传的附件 都是一个默认名字
	if fileName == "image.png" {
		fileName = GetFileName(uuid.New().String() + fileName)
	} else {
		fileName = GetFileName(fileName)
	}
	return fileName
}

// GetFileExt gets file extension
// GetFileExt 获取文件后缀
func GetFileExt(name string) string {
	return path.Ext(name)
}

// GetDatePath gets date save path
// GetDatePath 获取日期保存路径
func GetDatePath(timeFormat string) string {
	now := time.Now()
	if timeFormat == "" {
		timeFormat = "200601/02"
	}
	datePath := PathSuffixCheckAdd(now.Format(timeFormat), "/")
	return datePath
}

// IsContainExt determines if file extension is within the allowed range
// IsContainExt 判断文件后缀是否在允许范围内
func IsContainExt(t FileType, name string, uploadAllowExts []string) bool {
	ext := GetFileExt(name)
	ext = strings.ToUpper(ext)
	switch t {
	case ImageType:
		for _, allowExt := range uploadAllowExts {
			if strings.ToUpper(allowExt) == ext {
				return true
			}
		}
	}
	return false
}

// IsFileSizeAllowed determines if file size is allowed
// IsFileSizeAllowed 文件大小是否被允许
func IsFileSizeAllowed(t FileType, f multipart.File, uploadMaxSize int) bool {
	content, _ := io.ReadAll(f)
	size := len(content)
	switch t {
	case ImageType:
		if size >= uploadMaxSize*1024*1024 {
			return true
		}
	}
	return false
}

// IsPermission checks file permissions
// IsPermission 检查文件权限
func IsPermission(dst string) bool {
	_, err := os.Stat(dst)
	return os.IsPermission(err)
}

// IsExist determines if the given path exists
// IsExist 判断所给路径是否不存在
func IsExist(dst string) bool {
	_, err := os.Stat(dst) // os.Stat gets file info
	// os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// CreatePath creates path
// CreatePath 创建路径
func CreatePath(dst string, perm os.FileMode) error {
	dir := filepath.Dir(dst)
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return err
	}
	return nil
}

// GetExePath gets path of current execution file
// GetExePath 获取当前执行文件的路径
func GetExePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index]
}

// PathSuffixCheckAdd checks path suffix, adds it if not exists
// PathSuffixCheckAdd 检查路径后缀，如果没有则添加
func PathSuffixCheckAdd(path string, suffix string) string {
	if !strings.HasSuffix(path, suffix) {
		path = path + suffix
	}
	return path
}

// IsAbsPath determines if it is an absolute path
// IsAbsPath 判断是否为绝对路径
func IsAbsPath(path string) bool {
	if runtime.GOOS == "windows" {
		// Windows system
		// Windows系统
		if filepath.VolumeName(path) != "" {
			return true // If there is a drive letter, it is an absolute path
			// 如果有盘符，则为绝对路径
		}
		return filepath.IsAbs(path) // Check if it is an absolute path
		// 检查是否是绝对路径
	}
	// UNIX/Linux system
	// UNIX/Linux系统
	return filepath.IsAbs(path)
}

// GetAbsPath gets absolute path
// GetAbsPath 获取绝对路径
func GetAbsPath(path string, root string) (string, error) {
	if root != "" {
		root = PathSuffixCheckAdd(root, "/")
	}
	realPath := root + path
	// If it is already an absolute path, return directly
	// 如果本身就是绝对路径 就直接返回
	if !IsAbsPath(realPath) {
		pwdDir, _ := os.Getwd()
		realPath = PathSuffixCheckAdd(pwdDir, "/") + path
	}
	if IsExist(realPath) {
		return realPath, nil
	} else {
		return "", errors.New("file not exists")
	}
}

// CopyFile copies temporary file to target save path
// CopyFile 将临时文件复制到目标保存路径
// srcPath: absolute or relative path of source file
// srcPath: 源文件的绝对或相对路径
// destPath: full path of target save file (including file name)
// destPath: 目标保存文件的完整路径（包含文件名）
func CopyFile(srcPath, destPath string) error {
	// 1. Open source file
	// 1. 打开源文件
	sourceFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 2. Ensure target directory exists
	// 2. 确保目标目录存在
	// Recursively create directory, permissions set to 0754
	// 递归创建目录，权限设置为 0754
	if err := os.MkdirAll(filepath.Dir(destPath), 0754); err != nil {
		return err
	}

	// 3. Create target file
	// 3. 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 4. Perform copy operation
	// 4. 执行复制操作
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return nil
}
