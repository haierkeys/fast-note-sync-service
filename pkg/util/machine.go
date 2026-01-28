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

// GetMachineID gets unique identifier of the current machine
// GetMachineID 获取当前机器的唯一标识符
// Prioritize machineid library, fallback to motherboard serial number
// 优先使用 machineid 库，失败则尝试获取主板序列号
// return: machine ID string, returns empty string if all failed
// 返回值: 机器ID字符串，如果全部获取失败则返回空字符串
func GetMachineID() string {
	machineIDMutex.Lock()
	defer machineIDMutex.Unlock()

	if machineID != "" {
		return machineID
	}

	// 1. Try using machineid library
	// 1. 尝试使用 machineid 库
	id, err := machineid.ID()
	if err == nil && id != "" {
		machineID = id
		return machineID
	}

	// 2. Try getting motherboard serial number
	// 2. 尝试获取主板序列号
	id, err = getMotherboardID()
	if err == nil && id != "" {
		machineID = id
		return machineID
	}

	// 3. All failed, return empty string
	// 3. 全部失败，返回空字符串
	// Caller should determine if machine ID was successfully obtained based on the return value
	// 调用者应根据返回值判断是否成功获取机器ID
	return ""
}

// getMotherboardID gets the serial number of the motherboard
// getMotherboardID 获取主板序列号
func getMotherboardID() (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("wmic", "baseboard", "get", "serialnumber")
	case "linux":
		// Read file
		// 读取文件
		content, err := os.ReadFile("/sys/class/dmi/id/board_serial")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	case "darwin":
		cmd = exec.Command("ioreg", "-l") // Needs grep, a bit complex, simplified or empty here
		// ioreg -l | grep IOPlatformSerialNumber
		// Does not complete macOS complex parsing, simply return error to fallback
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

// parseSerialNumber parses the serial number from command output
// parseSerialNumber 从命令输出中解析序列号
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
