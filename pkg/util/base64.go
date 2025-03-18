package util

import (
	"encoding/base64"
)

// base64 加密
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// base64 解密
func base64Decode(s string) string {
	sByte, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return string(sByte)
	} else {
		return ""
	}
}
