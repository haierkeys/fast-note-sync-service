package convert

// Bool2Int converts a boolean to an integer
// Bool2Int 将布尔值转换为整数
// b: boolean value // 布尔值
// return: 1 if true, 0 if false // 返回值: true 返回 1，false 返回 0
func Bool2Int(b bool) int64 {
	var i int64
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}
