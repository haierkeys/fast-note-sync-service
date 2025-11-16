package util

import (
	"fmt"

	"github.com/denisbrodbeck/machineid"
)

// GetMachineID 获取当前机器的唯一标识符
// 返回值: 机器ID字符串，如果获取失败则返回空字符串
func GetMachineID() string {
	id, err := machineid.ID()
	if err != nil {
		fmt.Println("Failed to get machine ID")
		return ""
	}
	return id
}
