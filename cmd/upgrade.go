package cmd

import (
	"fmt"
	"os"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/upgrade"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade database schema to the latest version",
	Long: `Upgrade database schema to the latest version.

This command will check the current database version and apply all pending migrations.
It is safe to run this command multiple times - already applied migrations will be skipped.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 加载配置
		configPath, _ := cmd.Flags().GetString("config")
		if len(configPath) <= 0 {
			configPath = "config/config.yaml"
		}

		configRealpath, err := global.ConfigLoad(configPath)
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Loading config from: %s\n", configRealpath)

		// 初始化日志
		lg, err := logger.NewLogger(logger.Config{
			Level:      global.Config.Log.Level,
			File:       global.Config.Log.File,
			Production: global.Config.Log.Production,
		})
		if err != nil {
			fmt.Printf("Failed to init logger: %v\n", err)
			os.Exit(1)
		}
		global.Logger = lg

		// 初始化数据库
		global.DBEngine, err = dao.NewDBEngine(global.Config.Database)
		if err != nil {
			fmt.Printf("Failed to init database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Starting database upgrade...")

		// 执行升级
		if err := upgrade.Execute(); err != nil {
			fmt.Printf("Upgrade failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Database upgrade completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringP("config", "c", "", "config file path")
}
