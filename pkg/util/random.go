package util

import (
	"math/rand"
	"time"
)

// GenerateRandomNumber 生成一组指定范围内、不重复的随机整数切片
// start: 随机数的最小值
// end: 随机数的最大值
// count: 生成的随机数个数
// 返回值: 生成的随机数切片

func GenerateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}
	total := end - start
	// 这是一个打乱的序列 [0, 1, 5, 2, 4...]
	perm := rand.Perm(total)

	nums := make([]int, count)
	for i := 0; i < count; i++ {
		nums[i] = perm[i] + start
	}
	return nums
}

// InArray 检查整数是否在切片中（用于随机数生成）
// nums: 整数切片
// num: 待检查的整数
// 返回值: 如果在切片中返回true，否则返回false
func InArray(nums []int, num int) bool {
	for _, v := range nums {
		if v == num {
			return true
		}
	}
	return false
}

// GenerateRandomSingleNumber 生成单个随机数
// start: 随机数的最小值
// end: 随机数的最大值
// 返回值: 生成的随机数
func GenerateRandomSingleNumber(start int, end int) int {
	if end < start {
		return start
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(end-start) + start
}

// GetRandomString 生成指定长度的随机字符串
func GetRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		// 直接使用全局 rand，无需每次都 NewSource
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
