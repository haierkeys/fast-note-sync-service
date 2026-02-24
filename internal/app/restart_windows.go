//go:build windows

package app

import (
	"os"
	"os/exec"
)

// RestartProcess restarts the current process by starting a new one and exiting
// RestartProcess 通过启动新进程并退出来重启当前进程
func RestartProcess(argv0 string, args []string, env []string) error {
	cmd := exec.Command(argv0, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
