package cmd

import (
	"context"
	"fmt"
	"os"

	internalApp "github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func init() {
	var configPath string
	var username string
	var password string

	var resetPasswordCmd = &cobra.Command{
		Use:   "reset-password -u <username> -p <password> [-c config_file]",
		Short: "Reset a user's password by username",
		// 通过用户名重置用户密码，无需旧密码
		Run: func(cmd *cobra.Command, args []string) {
			if username == "" {
				bootstrapLogger.Error("username is required, use -u flag")
				os.Exit(1)
			}
			if password == "" {
				bootstrapLogger.Error("password is required, use -p flag")
				os.Exit(1)
			}

			// Load configuration
			// 加载配置
			if configPath == "" {
				// We can rely on default logic in further layers or set a default here
				// Use the same logic as run.go for consistency
				if fileurl.IsExist("config/config-dev.yaml") {
					configPath = "config/config-dev.yaml"
				} else if fileurl.IsExist("config.yaml") {
					configPath = "config.yaml"
				} else {
					configPath = "config/config.yaml"
				}
			}

			appConfig, configRealpath, err := internalApp.LoadConfig(configPath)
			if err != nil {
				bootstrapLogger.Error("failed to load config", zap.Error(err))
				os.Exit(1)
			}
			bootstrapLogger.Info("loading config", zap.String("path", configRealpath))

			// Initialize logger
			// 初始化日志
			lg, err := logger.NewLogger(logger.Config{
				Level:      appConfig.Log.Level,
				File:       appConfig.Log.File,
				Production: appConfig.Log.Production,
			})
			if err != nil {
				bootstrapLogger.Error("failed to init logger", zap.Error(err))
				os.Exit(1)
			}

			// Initialize database
			// 初始化数据库
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
				bootstrapLogger.Error("failed to init database", zap.Error(err))
				os.Exit(1)
			}

			// Initialize Dao and UserRepository
			// 初始化 Dao 和 UserRepository
			ctx := context.Background()
			daoObj := dao.New(db, ctx, dao.WithConfig(&dbConfig), dao.WithLogger(lg))
			userRepo := dao.NewUserRepository(daoObj)

			// Look up target user by username
			// 根据用户名查找目标用户
			user, err := userRepo.GetByUsername(ctx, username)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					fmt.Fprintf(os.Stderr, "Error: user '%s' not found\n", username)
				} else {
					fmt.Fprintf(os.Stderr, "Error: failed to query user: %v\n", err)
				}
				os.Exit(1)
			}

			// Generate password hash
			// 生成密码哈希
			hashedPassword, err := util.GeneratePasswordHash(password)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to generate password hash: %v\n", err)
				os.Exit(1)
			}

			// Update password
			// 更新密码
			if err := userRepo.UpdatePassword(ctx, hashedPassword, user.UID); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to update password: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Password for user '%s' (uid=%d) has been reset successfully.\n", username, user.UID)
		},
	}

	rootCmd.AddCommand(resetPasswordCmd)
	fs := resetPasswordCmd.Flags()
	fs.StringVarP(&configPath, "config", "c", "", "config file path (default: config/config.yaml)")
	fs.StringVarP(&username, "username", "u", "", "target username (required)")
	fs.StringVarP(&password, "password", "p", "", "new password (required)")
}
