package cmd

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	internalApp "github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type runFlags struct {
	dir     string // Project root directory // 项目根目录
	port    string // Startup port // 启动端口
	runMode string // Startup mode // 启动模式
	config  string // Specified configuration file path // 指定要使用的配置文件路径
}

func init() {
	runEnv := new(runFlags)

	var runCommand = &cobra.Command{
		Use:   "run [-c config_file] [-d working_dir] [-p port]",
		Short: "Run service",
		Run: func(cmd *cobra.Command, args []string) {
			if len(runEnv.dir) > 0 {
				err := os.Chdir(runEnv.dir)
				if err != nil {
					bootstrapLogger.Error("failed to change the current working directory", zap.Error(err))
				}
				bootstrapLogger.Info("working directory changed", zap.String("dir", runEnv.dir))
			}

			if len(runEnv.config) <= 0 {
				if fileurl.IsExist("config/config-dev.yaml") {
					runEnv.config = "config/config-dev.yaml"
				} else if fileurl.IsExist("config.yaml") {
					runEnv.config = "config.yaml"
				} else if fileurl.IsExist("config/config.yaml") {
					runEnv.config = "config/config.yaml"
				} else {

					bootstrapLogger.Warn("config file not found, creating default config")
					runEnv.config = "config/config.yaml"

					configDefault = strings.Replace(configDefault, "fast-note-sync-Auth-Token", util.GetRandomString(32), 1)

					if err := fileurl.CreatePath(runEnv.config, os.ModePerm); err != nil {
						bootstrapLogger.Error("config file auto create error", zap.Error(err))
						return
					}

					file, err := os.OpenFile(runEnv.config, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
					if err != nil {
						bootstrapLogger.Error("config file auto create error", zap.Error(err))
						return
					}
					defer file.Close()
					_, err = file.WriteString(configDefault)
					if err != nil {
						bootstrapLogger.Error("config file auto create writing error", zap.Error(err))
						return
					}
					bootstrapLogger.Info("config file auto create successfully", zap.String("path", runEnv.config))

				}
			}

			s, err := NewServer(runEnv)
			if err != nil {
				bootstrapLogger.Error("api service start err", zap.Error(err))
				return
			}

			go func() {

				w := watcher.New()

				// Set MaxEvents to 1 to receive at most 1 event in each listening cycle
				// 将 SetMaxEvents 设置为 1，以便在每个监听周期中至多接收 1 个事件
				// If SetMaxEvents is not set, all events are sent by default.
				// 如果没有设置 SetMaxEvents，默认情况下会发送所有事件。
				w.SetMaxEvents(1)

				// Only notify write events.
				// 只通知写入事件。
				w.FilterOps(watcher.Write)

				go func() {
					for {
						select {
						case event := <-w.Event:

							s.logger.Info("config watcher change", zap.String("event", event.Op.String()), zap.String("file", event.Path))
							s.sc.SendCloseSignal(nil)

							// Re-initialize server
							// 重新初始化 server
							s, err = NewServer(runEnv)
							if err != nil {
								bootstrapLogger.Error("service start err", zap.Error(err))
								continue
							}

						case err := <-w.Error:
							s.logger.Error("config watcher error", zap.Error(err))
						case <-w.Closed:
							bootstrapLogger.Info("config watcher closed")
						}
					}
				}()

				// Watch config.yaml file
				// 监听 config.yaml 文件
				if err := w.Add(runEnv.config); err != nil {
					s.logger.Error("config watcher file error", zap.Error(err))
				}

				// Start watching
				// 启动监听
				if err := w.Start(time.Second * 5); err != nil {
					s.logger.Error("config watcher start error", zap.Error(err))
				}
			}()

			quit1 := make(chan os.Signal, 1)
			signal.Notify(quit1, syscall.SIGINT, syscall.SIGTERM)

			select {
			case <-quit1:
				s.logger.Info("Received shutdown signal, initiating graceful shutdown...")
				s.sc.SendCloseSignal(nil)
			case newBinaryPath := <-s.GetApp().UpgradeSignal:
				s.logger.Info("Received upgrade signal, starting smooth restart...", zap.String("newBinary", newBinaryPath))

				// 1. Graceful shutdown (close servers, release ports)
				s.sc.SendCloseSignal(nil)
				if err := s.sc.WaitClosed(); err != nil {
					s.logger.Error("Shutdown failed before upgrade", zap.Error(err))
				}

				// 2. Perform file replacement
				currentBinary, _ := os.Executable()
				oldBinary := currentBinary + ".old"
				_ = os.Remove(oldBinary)
				if err := util.MoveFile(currentBinary, oldBinary); err != nil {
					s.logger.Error("Failed to backup current binary", zap.Error(err))
					return
				}
				if err := util.MoveFile(newBinaryPath, currentBinary); err != nil {
					s.logger.Error("Failed to replace binary", zap.Error(err))
					// Try to restore?
					_ = util.MoveFile(oldBinary, currentBinary)
					return
				}
				if err := os.Chmod(currentBinary, 0755); err != nil {
					s.logger.Error("Failed to set executable permission", zap.Error(err))
				}

				// 2.1 Cleanup temp directory (where the tar.gz and temporary binary were)
				// 清理临时目录（包含压缩包和临时的二进制文件）
				tempDir := filepath.Dir(newBinaryPath)
				if err := os.RemoveAll(tempDir); err != nil {
					s.logger.Warn("Failed to cleanup upgrade temp directory", zap.String("path", tempDir), zap.Error(err))
				}

				// 3. Restart
				env := os.Environ()
				args := os.Args
				if err := internalApp.RestartProcess(currentBinary, args, env); err != nil {
					s.logger.Error("Failed to restart process", zap.Error(err))
				}
				return // Should not reach here on Linux due to syscall.Exec
			}

			// Wait for all shutdown handlers to complete (including App Container graceful shutdown)
			// 等待所有关闭处理器完成（包括 App Container 的优雅关闭）
			if err := s.sc.WaitClosed(); err != nil {
				s.logger.Error("Shutdown completed with error", zap.Error(err))
			} else {
				s.logger.Info("Service has been shut down gracefully.")
			}

		},
	}

	rootCmd.AddCommand(runCommand)
	fs := runCommand.Flags()
	fs.StringVarP(&runEnv.dir, "dir", "d", "", "run dir")
	fs.StringVarP(&runEnv.port, "port", "p", "", "run port")
	fs.StringVarP(&runEnv.runMode, "mode", "m", "", "run mode")
	fs.StringVarP(&runEnv.config, "config", "c", "", "config file")

}
