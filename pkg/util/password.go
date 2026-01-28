package util

import (
	"golang.org/x/crypto/bcrypt"
)

// GeneratePasswordHash generates bcrypt hash of a password
// GeneratePasswordHash 生成密码的bcrypt哈希值
// password: original password string // 原始密码字符串
// return: hashed password string, and possible error info // 返回值: 哈希后的密码字符串，以及可能的错误信息
func GeneratePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CheckPasswordHash verifies whether password matches the hash
// CheckPasswordHash 验证密码与哈希值是否匹配
// hash: stored hash value // 存储的哈希值
// password: password to be verified // 待验证的密码
// return: true if password matches, false otherwise // 返回值: 如果密码匹配返回true，否则返回false
func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
