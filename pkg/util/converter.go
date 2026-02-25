package util

import (
	"fmt"
	"strconv"
	"strings"
)

// StrToMap converts string to map
// StrToMap 将字符串转换为map
// str: string in the format of "key=value,key=value" // 格式为"key=value,key=value"的字符串
// return: converted map // 返回值: 转换后的map
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

// StrToInt converts string to integer
// StrToInt 将字符串转换为整数
// str: string to be converted // 待转换的字符串
// return: converted integer, or 0 if conversion fails // 返回值: 转换后的整数，如果转换失败返回0
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

// IntSliceToStringSlice converts integer slice to string slice
// IntSliceToStringSlice 将整数切片转换为字符串切片
// intSlice: integer slice // 整数切片
// return: string slice // 返回值: 字符串切片
func IntSliceToStringSlice(intSlice []int) []string {
	stringSlice := make([]string, len(intSlice))
	for i, v := range intSlice {
		stringSlice[i] = strconv.Itoa(v)
	}
	return stringSlice
}

// StringToInt64 converts string to int64
// StringToInt64 将字符串转换为int64
// s: string to be converted // 待转换的字符串
// return: converted int64 value // 返回值: 转换后的int64值
func StringToInt64(s string) int64 {
	result, _ := strconv.ParseInt(s, 10, 64)
	return result
}

// ParseSize parses size string like "128MB", "512KB", "1024B" to bytes
// ParseSize 将大小字符串（如 "128MB", "512KB", "1024B"）解析为字节数
func ParseSize(sizeStr string, defaultSize int64) int64 {
	if sizeStr == "" {
		return defaultSize
	}

	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	var multiplier int64 = 1

	if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "B") {
		multiplier = 1
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	size, err := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)
	if err != nil || size <= 0 {
		return defaultSize
	}

	return size * multiplier
}

// IntSliceToStrSlice converts integer slice to string slice (another implementation)
// IntSliceToStrSlice 将整数切片转换为字符串切片（另一种实现）
// list: integer slice // 整数切片
// return: string slice // 返回值: 字符串切片
func IntSliceToStrSlice(list []int) []string {
	strlist := make([]string, 0)
	for _, i := range list {
		strlist = append(strlist, fmt.Sprintf("%d", i))
	}
	return strlist
}
