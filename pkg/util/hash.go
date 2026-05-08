package util

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"unicode/utf16"
)

// EncodeMD5 performs MD5 encoding on a string
// EncodeMD5 对字符串进行MD5编码
// str: string to be encoded
// str: 待编码的字符串
// return: MD5 encoded 32-bit hexadecimal string
// 返回值: MD5编码后的32位十六进制字符串
func EncodeMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// EncodeHash32 performs 32-bit hash encoding on a string
// EncodeHash32 对字符串进行 32 位哈希编码
func EncodeHash32(content string) string {
	// Convert string to rune slice, then to UTF-16 code units (consistent with JS internal representation)
	// 首先将字符串转为 rune 切片，再转为 UTF-16 code units（与 JS 的内部表示一致）
	runes := []rune(content)
	utf16Units := utf16.Encode(runes) // []uint16
	var hash int32 = 0
	for _, u := range utf16Units {
		char := int32(u) // Consistent with 16-bit value returned by JS charCodeAt // 与 JS charCodeAt 返回的 16-bit 值一致
		hash = (hash << 5) - hash + char
		// int32 will automatically overflow, equivalent to JS 32-bit bitwise operation result
		// int32 会自动溢出，等价于 JS 的 32-bit 位运算结果
	}
	return strconv.Itoa(int(hash))
}

const (
	// FileHashThreshold defines the size threshold (100MB) above which partial hashing is used
	// FileHashThreshold 定义触发分段哈希的阈值 (100MB)
	FileHashThreshold = 100 * 1024 * 1024
	// FileHashSliceSize defines the size of slices taken from the beginning and end of large files (50MB)
	// FileHashSliceSize 定义大文件分段哈希时首尾读取的大小 (50MB)
	FileHashSliceSize = 50 * 1024 * 1024
)

// EncodeHash32Bytes performs 32-bit hash encoding on raw bytes.
// If the data exceeds 100MB, it only hashes the first 50MB and last 50MB.
// EncodeHash32Bytes 对原始字节进行 32 位哈希编码。
// 如果数据超过 100MB，则仅计算前 50MB 和后 50MB 的哈希。
func EncodeHash32Bytes(data []byte) string {
	size := len(data)
	var hash int32 = 0

	if size <= FileHashThreshold {
		// Small data: full hash // 小数据：全量哈希
		for _, b := range data {
			hash = (hash << 5) - hash + int32(b)
		}
	} else {
		// Large data: hash first 50MB + last 50MB // 大数据：哈希前 50MB + 后 50MB
		// Hash first 50MB
		for i := 0; i < FileHashSliceSize; i++ {
			hash = (hash << 5) - hash + int32(data[i])
		}
		// Hash last 50MB
		for i := size - FileHashSliceSize; i < size; i++ {
			hash = (hash << 5) - hash + int32(data[i])
		}
	}
	return strconv.Itoa(int(hash))
}
