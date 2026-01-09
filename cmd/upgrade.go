package cmd

import (
	"fmt"
	"os"

	"github.com/haierkeys/fast-note-sync-service/global"
	internalApp "github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/upgrade"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade legacy database schema and other data to the latest version",
	Long: `Upgrade legacy database schema and other data to the latest version.

This command will check the current database version and apply all pending migrations.
It is safe to run this command multiple times - already applied migrations will be skipped.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 加载配置
		configPath, _ := cmd.Flags().GetString("config")
		if len(configPath) <= 0 {
			configPath = "config/config.yaml"
		}

		// 使用 LoadConfig 直接加载配置到 AppConfig
		appConfig, configRealpath, err := internalApp.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Loading config from: %s\n", configRealpath)

		// 初始化日志
		lg, err := logger.NewLogger(logger.Config{
			Level:      appConfig.Log.Level,
			File:       appConfig.Log.File,
			Production: appConfig.Log.Production,
		})
		if err != nil {
			fmt.Printf("Failed to init logger: %v\n", err)
			os.Exit(1)
		}

		// 初始化数据库（使用注入的配置）
		dbConfig := dao.DatabaseConfig{
			Type:            appConfig.Database.Type,
			Path:            appConfig.Database.Path,
			UserName:        appConfig.Database.UserName,
			Password:        appConfig.Database.Password,
			Host:            appConfig.Database.Host,
			Name:            appConfig.Database.Name,
			TablePrefix:     appConfig.Database.TablePrefix,
			AutoMigrate:     appConfig.Database.AutoMigrate,
			Charset:         appConfig.Database.Charset,
			ParseTime:       appConfig.Database.ParseTime,
			MaxIdleConns:    appConfig.Database.MaxIdleConns,
			MaxOpenConns:    appConfig.Database.MaxOpenConns,
			ConnMaxLifetime: appConfig.Database.ConnMaxLifetime,
			ConnMaxIdleTime: appConfig.Database.ConnMaxIdleTime,
			RunMode:         appConfig.Server.RunMode,
		}

		db, err := dao.NewDBEngineWithConfig(dbConfig, lg)
		if err != nil {
			fmt.Printf("Failed to init database: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Starting database upgrade...")

		// 执行升级
		if err := upgrade.Execute(
			db,
			lg,
			global.Version,
			appConfig.Database.Path,
			appConfig.Database.Type,
		); err != nil {
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
