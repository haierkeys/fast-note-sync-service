package util

// GetIndexSlice 获取切片元素的索引
// arr: 待查找的切片
// val: 要查找的值
// 返回值: 元素的索引，如果不存在返回-1
func GetIndexSlice(arr []string, val string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}

// InSlice 判断元素是否在切片中（泛型版本）
// 参数:
//
//	slice - 切片
//	item - 要查找的元素
//
// 返回值: bool - 存在返回true，否则返回false
func InSlice[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Inarray 判断元素是否在切片中
// arr: 待查找的切片
// val: 要查找的值
// 返回值: 如果元素在切片中返回true，否则返回false
func Inarray(arr []string, val string) bool {
	return GetIndexSlice(arr, val) >= 0
}

// ArrayUnique 移除切片中的重复元素
// arr: 原始切片
// 返回值: 去重后的新切片
func ArrayUnique(arr []string) []string {
	result := make([]string, 0)
	m := make(map[string]bool)
	for _, v := range arr {
		if !m[v] {
			m[v] = true
			result = append(result, v)
		}
	}
	return result
}

// RemoveDuplicate 移除字符串切片中的重复元素（另一种实现）
// strSlice: 原始字符串切片
// 返回值: 去重后的字符串切片
func RemoveDuplicate(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// IntersectionInt 计算两个整数切片的交集
// a: 第一个整数切片
// b: 第二个整数切片
// 返回值: 两个切片的交集
func IntersectionInt(a, b []int) []int {
	hash := make(map[int]struct{})
	for _, v := range a {
		hash[v] = struct{}{}
	}
	result := make([]int, 0)
	for _, v := range b {
		if _, ok := hash[v]; ok {
			result = append(result, v)
		}
	}
	return result
}
