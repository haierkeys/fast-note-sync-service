package util

import (
	"crypto/md5"
	"encoding/hex"
)

// EncodeMD5 对字符串进行MD5编码
// str: 待编码的字符串
// 返回值: MD5编码后的32位十六进制字符串
func EncodeMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
