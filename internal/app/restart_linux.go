//go:build linux

package app

import (
	"syscall"
)

// RestartProcess restarts the current process using syscall.Exec
// RestartProcess 使用 syscall.Exec 重启当前进程
func RestartProcess(argv0 string, args []string, env []string) error {
	return syscall.Exec(argv0, args, env)
}
