package util

import (
	"regexp"
)

// IsValidEmail verifies if the email format is correct
// IsValidEmail 验证邮箱格式是否正确
// email: email string to be verified
// email: 待验证的邮箱字符串
// return: true if email format is correct, false otherwise
// 返回值: 如果邮箱格式正确返回true，否则返回false
func IsValidEmail(email string) bool {
	// Simple email format validation regular expression
	// 简单的邮箱格式验证正则表达式
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// IsValidUsername verifies if the username format is correct
// IsValidUsername 验证用户名格式是否正确
// username: username to be verified
// username: 待验证的用户名
// return: true if username format is correct, false otherwise
// 返回值: 如果用户名格式正确返回true，否则返回false
func IsValidUsername(username string) bool {
	// Username format: letters, numbers, underscores, length 3-20
	// 用户名格式：字母、数字、下划线，长度3-20
	pattern := `^[a-zA-Z0-9_]{3,20}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(username)
}
