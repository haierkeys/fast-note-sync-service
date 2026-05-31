package util

import (
	"io"
	"os"
	"path/filepath"

	"github.com/yeka/zip"
)

// Zip compresses files or directories into a zip file
// source: path to file or directory
// target: path to output zip file
func Zip(source, target string) error {
	return ZipWithPassword(source, target, "")
}

// ZipWithPassword compresses files or directories into a zip file with password
func ZipWithPassword(source, target, password string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	if _, err := os.Stat(source); err != nil {
		return err
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from source
		// 获取相对于 source 的路径
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Skip root directory itself
		// 跳过根目录本身
		if relPath == "." {
			return nil
		}

		// Use / as separator in zip file (Standard requirements)
		// ZIP 文件内必须使用 / 作为路径分隔符
		relPath = filepath.ToSlash(relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Use relative path as filename in zip
		// 使用相对路径作为压缩包内的文件名
		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		var writer io.Writer
		if password != "" && !info.IsDir() {
			// Set password and encryption method
			// 设置密码和加密方法
			header.SetPassword(password)
			// StandardEncryption is more compatible with Windows built-in ZIP tool
			// StandardEncryption 对 Windows 自带 ZIP 工具兼容性更好
			header.SetEncryptionMethod(zip.StandardEncryption)
			writer, err = archive.CreateHeader(header)
		} else {
			writer, err = archive.CreateHeader(header)
		}

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// ZipBytes creates a zip archive from a map of filenames and their contents (bytes)
func ZipBytes(files map[string][]byte, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	for name, content := range files {
		writer, err := archive.Create(name)
		if err != nil {
			return err
		}
		_, err = writer.Write(content)
		if err != nil {
			return err
		}
	}

	return nil
}
