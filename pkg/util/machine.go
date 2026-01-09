package util

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/denisbrodbeck/machineid"
)

var (
	machineID      string
	machineIDMutex sync.Mutex
)

// GetMachineID 获取当前机器的唯一标识符
// 优先使用 machineid 库，失败则尝试获取主板序列号
// 返回值: 机器ID字符串，如果全部获取失败则返回空字符串
func GetMachineID() string {
	machineIDMutex.Lock()
	defer machineIDMutex.Unlock()

	if machineID != "" {
		return machineID
	}

	// 1. 尝试使用 machineid 库
	id, err := machineid.ID()
	if err == nil && id != "" {
		machineID = id
		return machineID
	}

	// 2. 尝试获取主板序列号
	id, err = getMotherboardID()
	if err == nil && id != "" {
		machineID = id
		return machineID
	}

	// 3. 全部失败，返回空字符串
	// 调用者应根据返回值判断是否成功获取机器ID
	return ""
}

func getMotherboardID() (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("wmic", "baseboard", "get", "serialnumber")
	case "linux":
		// 读取文件
		content, err := os.ReadFile("/sys/class/dmi/id/board_serial")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	case "darwin":
		cmd = exec.Command("ioreg", "-l") // 需要配合 grep，复杂一点，这里先简化处理或留空
		// ioreg -l | grep IOPlatformSerialNumber
		// 暂不完整实现 macOS 复杂解析，简单返回 error 走 fallback
		return "", errors.New("not implemented for darwin")
	default:
		return "", errors.New("unsupported os")
	}

	if cmd != nil {
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return parseSerialNumber(string(out)), nil
	}

	return "", errors.New("unknown error")
}

func parseSerialNumber(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.EqualFold(line, "SerialNumber") {
			continue
		}
		return line
	}
	return ""
}
