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
	// 首先将字符串转为 rune 切片，再转为 UTF-16 code units（与 JS 的内部表示一致）
	runes := []rune(content)
	utf16Units := utf16.Encode(runes) // []uint16
	var hash int32 = 0
	for _, u := range utf16Units {
		char := int32(u) // 与 JS charCodeAt 返回的 16-bit 值一致
		hash = (hash << 5) - hash + char
		// int32 会自动溢出，等价于 JS 的 32-bit 位运算结果
	}
	return strconv.Itoa(int(hash))
}
