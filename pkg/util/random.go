package util

import (
	"math/rand"
	"time"
)

// GenerateRandomNumber generates a slice of unique random integers within specified range
// GenerateRandomNumber 生成一组指定范围内、不重复的随机整数切片
// start: minimum value of random number
// start: 随机数的最小值
// end: maximum value of random number
// end: 随机数的最大值
// count: number of random numbers to generate
// count: 生成的随机数个数
// return: generated random number slice
// 返回值: 生成的随机数切片

func GenerateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}
	total := end - start
	// This is a shuffled sequence [0, 1, 5, 2, 4...]
	// 这是一个打乱的序列 [0, 1, 5, 2, 4...]
	perm := rand.Perm(total)

	nums := make([]int, count)
	for i := 0; i < count; i++ {
		nums[i] = perm[i] + start
	}
	return nums
}

// InArray checks whether an integer is in a slice (used for random number generation)
// InArray 检查整数是否在切片中（用于随机数生成）
// nums: integer slice
// nums: 整数切片
// num: integer to be checked
// num: 待检查的整数
// return: true if in slice, false otherwise
// 返回值: 如果在切片中返回true，否则返回false
func InArray(nums []int, num int) bool {
	for _, v := range nums {
		if v == num {
			return true
		}
	}
	return false
}

// GenerateRandomSingleNumber generates a single random number
// GenerateRandomSingleNumber 生成单个随机数
// start: minimum value of random number
// start: 随机数的最小值
// end: maximum value of random number
// end: 随机数的最大值
// return: generated random number
// 返回值: 生成的随机数
func GenerateRandomSingleNumber(start int, end int) int {
	if end < start {
		return start
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(end-start) + start
}

// GetRandomString generates random string of specified length
// GetRandomString 生成指定长度的随机字符串
func GetRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		// Use global rand directly, no need to NewSource every time
		// 直接使用全局 rand，无需每次都 NewSource
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
