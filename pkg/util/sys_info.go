package util

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// GetOSPrettyName gets a more readable and detailed OS name and version
// GetOSPrettyName 获取更具可读性和详细的操作系统名称及版本
func GetOSPrettyName() string {
	switch runtime.GOOS {
	case "linux":
		return getLinuxPrettyName()
	case "windows":
		return getWindowsVersion()
	case "darwin":
		return getMacOSVersion()
	default:
		return runtime.GOOS
	}
}

// getLinuxPrettyName reads /etc/os-release to get PRETTY_NAME
// getLinuxPrettyName 读取 /etc/os-release 获取 PRETTY_NAME
func getLinuxPrettyName() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			name := strings.TrimPrefix(line, "PRETTY_NAME=")
			name = strings.Trim(name, "\"")
			return name
		}
	}
	return "Linux"
}

// getWindowsVersion executes 'cmd /c ver' to get Windows version
// getWindowsVersion 执行 'cmd /c ver' 获取 Windows 版本
func getWindowsVersion() string {
	cmd := exec.Command("cmd", "/c", "ver")
	out, err := cmd.Output()
	if err != nil {
		return "Windows"
	}
	return strings.TrimSpace(string(out))
}

// getMacOSVersion executes 'sw_vers -productVersion' to get macOS version
// getMacOSVersion 执行 'sw_vers -productVersion' 获取 macOS 版本
func getMacOSVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	out, err := cmd.Output()
	if err != nil {
		return "macOS"
	}
	return "macOS " + strings.TrimSpace(string(out))
}
