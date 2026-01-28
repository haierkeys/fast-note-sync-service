package util

// GetIndexSlice gets the index of a slice element
// GetIndexSlice 获取切片元素的索引
// arr: slice to search
// arr: 待查找的切片
// val: value to search for
// val: 要查找的值
// return: index of the element, or -1 if not found
// 返回值: 元素的索引，如果不存在返回-1
func GetIndexSlice(arr []string, val string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}

// InSlice determines whether an element is in a slice (generic version)
// InSlice 判断元素是否在切片中（泛型版本）
// slice: the slice // 切片
// item: the element to find // 要查找的元素
// return: bool - true if exists, false otherwise // 返回值: bool - 存在返回true，否则返回false
func InSlice[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Inarray determines whether an element is in a slice
// Inarray 判断元素是否在切片中
// arr: slice to search // 待查找的切片
// val: value to search for // 要查找的值
// return: true if element is in the slice, false otherwise // 返回值: 如果元素在切片中返回true，否则返回false
func Inarray(arr []string, val string) bool {
	return GetIndexSlice(arr, val) >= 0
}

// ArrayUnique removes duplicate elements from a slice
// ArrayUnique 移除切片中的重复元素
// arr: original slice // 原始切片
// return: new slice without duplicates // 返回值: 去重后的新切片
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

// RemoveDuplicate removes duplicate elements from a string slice (another implementation)
// RemoveDuplicate 移除字符串切片中的重复元素（另一种实现）
// strSlice: original string slice // 原始字符串切片
// return: string slice without duplicates // 返回值: 去重后的字符串切片
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

// IntersectionInt calculates intersection of two integer slices
// IntersectionInt 计算两个整数切片的交集
// a: first integer slice // 第一个整数切片
// b: second integer slice // 第二个整数切片
// return: intersection of two slices // 返回值: 两个切片的交集
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
