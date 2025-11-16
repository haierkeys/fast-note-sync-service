package util

import (
	"math/rand"
	"time"
)

// GenerateRandomNumber 生成指定范围的随机数切片
// start: 随机数的最小值
// end: 随机数的最大值
// count: 生成的随机数个数
// 返回值: 生成的随机数切片
func GenerateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}

	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for len(nums) < count {
		num := r.Intn(end-start) + start
		// 检查是否已存在
		if !InArray(nums, num) {
			nums = append(nums, num)
		}
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

// GenerateRandom 生成指定长度的随机字符串
// n: 随机字符串的长度
// 返回值: 生成的随机字符串
func GenerateRandom(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
