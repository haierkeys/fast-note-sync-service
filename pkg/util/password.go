package util

import (
	"golang.org/x/crypto/bcrypt"
)

// GeneratePasswordHash 生成密码的bcrypt哈希值
// password: 原始密码字符串
// 返回值: 哈希后的密码字符串，以及可能的错误信息
func GeneratePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CheckPasswordHash 验证密码与哈希值是否匹配
// hash: 存储的哈希值
// password: 待验证的密码
// 返回值: 如果密码匹配返回true，否则返回false
func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
