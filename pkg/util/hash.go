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
	// 使用简单的哈希函数生成哈希值
	var hash int32 = 0
	for i := 0; i < len(content); i++ {
		char := int32(content[i])
		hash = (hash << 5) - hash + char
		hash &= hash
	}
	return strconv.Itoa(int(hash))
}
