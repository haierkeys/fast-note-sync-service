package util

import (
	"fmt"
	"strconv"
	"strings"
)

// StrToMap 将字符串转换为map
// str: 格式为"key=value,key=value"的字符串
// 返回值: 转换后的map
func StrToMap(str string) map[string]string {
	result := make(map[string]string)
	if str == "" {
		return result
	}

	strArr := strings.Split(str, ",")
	for _, item := range strArr {
		kv := strings.Split(item, "=")
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result
}

// StrToInt 将字符串转换为整数
// str: 待转换的字符串
// 返回值: 转换后的整数，如果转换失败返回0
func StrToInt(str string) int {
	if str == "" {
		return 0
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

// IntSliceToStringSlice 将整数切片转换为字符串切片
// intSlice: 整数切片
// 返回值: 字符串切片
func IntSliceToStringSlice(intSlice []int) []string {
	stringSlice := make([]string, len(intSlice))
	for i, v := range intSlice {
		stringSlice[i] = strconv.Itoa(v)
	}
	return stringSlice
}

// StringToInt64 将字符串转换为int64
// s: 待转换的字符串
// 返回值: 转换后的int64值
func StringToInt64(s string) int64 {
	result, _ := strconv.ParseInt(s, 10, 64)
	return result
}

// IntSliceToStrSlice 将整数切片转换为字符串切片（另一种实现）
// list: 整数切片
// 返回值: 字符串切片
func IntSliceToStrSlice(list []int) []string {
	strlist := make([]string, 0)
	for _, i := range list {
		strlist = append(strlist, fmt.Sprintf("%d", i))
	}
	return strlist
}
