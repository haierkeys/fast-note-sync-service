package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type runFlags struct {
	dir     string // 项目根目录
	port    string // 启动端口
	runMode string // 启动模式
	config  string // 指定要使用的配置文件路径
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

				// 将 SetMaxEvents 设置为 1，以便在每个监听周期中至多接收 1 个事件
				// 如果没有设置 SetMaxEvents，默认情况下会发送所有事件。
				w.SetMaxEvents(1)

				// 只通知重命名和移动事件。
				w.FilterOps(watcher.Write)

				go func() {
					for {
						select {
						case event := <-w.Event:

							s.logger.Info("config watcher change", zap.String("event", event.Op.String()), zap.String("file", event.Path))
							s.sc.SendCloseSignal(nil)

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

				// 监听 config.yaml 文件
				if err := w.Add(runEnv.config); err != nil {
					s.logger.Error("config watcher file error", zap.Error(err))
				}

				// 启动监听
				if err := w.Start(time.Second * 5); err != nil {
					s.logger.Error("config watcher start error", zap.Error(err))
				}
			}()

			quit1 := make(chan os.Signal, 1)
			signal.Notify(quit1, syscall.SIGINT, syscall.SIGTERM)
			<-quit1
			s.logger.Info("Received shutdown signal, initiating graceful shutdown...")
			s.sc.SendCloseSignal(nil)

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
