package util

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

// EncodeMD5 对字符串进行MD5编码
// str: 待编码的字符串
// 返回值: MD5编码后的32位十六进制字符串
func EncodeMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func EncodeHash32(content string) string {
	var hash int32 = 0
	// 将 string 转为 UTF-16 rune 数组以匹配 JS 的处理方式
	runes := []rune(content)
	for i := 0; i < len(runes); i++ {
		char := int32(runes[i])
		hash = (hash << 5) - hash + char
		// Go 的 int32 会自动溢出处理，无需额外操作
	}
	return strconv.Itoa(int(hash))
}
