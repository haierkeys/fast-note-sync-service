package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	internalApp "github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/routers"
	"github.com/haierkeys/fast-note-sync-service/internal/task"
	"github.com/haierkeys/fast-note-sync-service/internal/upgrade"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/safe_close"
	"github.com/haierkeys/fast-note-sync-service/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	validatorV10 "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// defaultSecretKeys 定义需要检测的默认密钥列表
var defaultSecretKeys = []string{
	"6666",
	"fast-note-sync-Auth-Token",
	"",
}

// DefaultShutdownTimeout 默认关闭超时时间
const DefaultShutdownTimeout = 30 * time.Second

type Server struct {
	logger            *zap.Logger                // non-nil logger.
	config            *internalApp.AppConfig     // 应用配置（注入的依赖）
	db                *gorm.DB                   // 数据库连接
	ut                *ut.UniversalTranslator    // 翻译器
	httpServer        *http.Server
	privateHttpServer *http.Server
	sc                *safe_close.SafeClose
	app               *internalApp.App // App Container
}

// checkSecurityConfig 检查安全配置，如果使用默认密钥则输出警告
func checkSecurityConfigWithConfig(cfg *internalApp.AppConfig, lg *zap.Logger) {
	isDefault := false
	for _, key := range defaultSecretKeys {
		if cfg.Security.AuthTokenKey == key {
			isDefault = true
			break
		}
	}

	if isDefault {
		// 输出到控制台
		fmt.Println()
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("⚠️  SECURITY WARNING: Using default secret key!")
		fmt.Println()
		fmt.Println("Please modify 'security.auth-token-key' in config.yaml")
		fmt.Println("Generate a secure key with:")
		fmt.Println("  openssl rand -base64 32")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()

		// 记录到日志
		if lg != nil {
			lg.Warn("Using default secret key - please change security.auth-token-key in config.yaml")
		}
	}
}

func NewServer(runEnv *runFlags) (*Server, error) {

	// 使用 LoadConfig 直接加载配置到 AppConfig
	appConfig, configRealpath, err := internalApp.LoadConfig(runEnv.config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 确定运行模式
	runMode := runEnv.runMode
	if len(runMode) <= 0 {
		runMode = appConfig.Server.RunMode
	}

	if len(runMode) > 0 {
		gin.SetMode(runMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		config: appConfig,
		sc:     safe_close.NewSafeClose(),
	}

	// 初始化日志器（使用注入的配置）
	if err := initLoggerWithConfig(s, appConfig); err != nil {
		return nil, fmt.Errorf("initLogger: %w", err)
	}

	// 检查安全配置（使用注入的配置）
	checkSecurityConfigWithConfig(appConfig, s.logger)

	// 初始化存储目录（使用注入的配置）
	if err := initStorageWithConfig(appConfig); err != nil {
		return nil, fmt.Errorf("initStorage: %w", err)
	}

	// 初始化数据库（使用注入的配置）
	db, err := initDatabaseWithConfig(appConfig, s.logger)
	if err != nil {
		return nil, fmt.Errorf("initDatabase: %w", err)
	}
	s.db = db

	// 初始化 App Container（直接使用 AppConfig）
	app, err := internalApp.NewApp(appConfig, s.logger, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create app container: %w", err)
	}
	s.app = app

	// 自动执行迁移任务（使用注入的配置）
	if err := upgrade.Execute(
		db,
		s.logger,
		global.Version,
		appConfig.Database.Path,
		appConfig.Database.Type,
	); err != nil {
		return nil, fmt.Errorf("upgrade.Execute: %w", err)
	}

	// 初始化验证器
	uni, err := initValidatorWithLogger(s.logger)
	if err != nil {
		return nil, fmt.Errorf("initValidator: %w", err)
	}
	s.ut = uni

	validator.RegisterCustom()

	// 启动调度器
	initScheduler(s)

	banner := `
    ______           __     _   __      __          _____
   / ____/___ ______/ /_   / | / /___  / /____     / ___/__  ______  _____
  / /_  / __  / ___/ __/  /  |/ / __ \/ __/ _ \    \__ \/ / / / __ \/ ___/
 / __/ / /_/ (__  ) /_   / /|  / /_/ / /_/  __/   ___/ / /_/ / / / / /__
/_/    \__,_/____/\__/  /_/ |_/\____/\__/\___/   /____/\__, /_/ /_/\___/
                                                      /____/              `
	s.logger.Warn(fmt.Sprintf("%s\n\n%s v%s\nGit: %s\nBuildTime: %s\n", banner, global.Name, global.Version, global.GitTag, global.BuildTime))

	s.logger.Warn("config loaded", zap.String("path", configRealpath))

	// 启动 HTTP API 服务器
	if httpAddr := appConfig.Server.HttpPort; len(httpAddr) > 0 {
		s.logger.Warn("api_router", zap.String("config.server.HttpPort", appConfig.Server.HttpPort))
		s.httpServer = &http.Server{
			Addr:           appConfig.Server.HttpPort,
			Handler:        routers.NewRouter(frontendFiles, s.app, s.ut),
			ReadTimeout:    time.Duration(appConfig.Server.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(appConfig.Server.WriteTimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.sc.Attach(func(done func(), closeSignal <-chan struct{}) {
			defer done()
			errChan := make(chan error, 1)
			go func() {
				errChan <- s.httpServer.ListenAndServe()
			}()
			select {
			case err := <-errChan:
				s.logger.Error("api service err", zap.Error(err))
				s.sc.SendCloseSignal(err)
			case <-closeSignal:

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// 停止HTTP服务器
				if err := s.httpServer.Shutdown(ctx); err != nil {
					s.logger.Error("api service shutdown error", zap.Error(err))
				}

				// _ = s.httpServer.Close()
			}
		})
	}

	if httpAddr := appConfig.Server.PrivateHttpListen; len(httpAddr) > 0 {

		s.logger.Info("api_router", zap.String("config.server.PrivateHttpListen", appConfig.Server.PrivateHttpListen))
		s.privateHttpServer = &http.Server{
			Addr:           appConfig.Server.PrivateHttpListen,
			Handler:        routers.NewPrivateRouterWithLogger(appConfig.Server.RunMode, s.logger),
			ReadTimeout:    time.Duration(appConfig.Server.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(appConfig.Server.WriteTimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		s.sc.Attach(func(done func(), closeSignal <-chan struct{}) {
			defer done()
			errChan := make(chan error, 1)
			go func() {
				errChan <- s.privateHttpServer.ListenAndServe()
			}()
			select {
			case err := <-errChan:
				s.logger.Error("private api service err", zap.Error(err))
				s.sc.SendCloseSignal(err)
			case <-closeSignal:

				// _ = s.httpServer.Close()

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// 停止HTTP服务器
				if err := s.privateHttpServer.Shutdown(ctx); err != nil {
					s.logger.Error("private api service shutdown error", zap.Error(err))
				}
			}
		})
	}

	// 注册 App Container 的优雅关闭（使用 Shutdown 方法）
	s.sc.Attach(func(done func(), closeSignal <-chan struct{}) {
		defer done()
		<-closeSignal
		if s.app != nil {
			// 使用带超时的优雅关闭
			ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
			defer cancel()

			if err := s.app.Shutdown(ctx); err != nil {
				s.logger.Error("failed to shutdown app container", zap.Error(err))
			} else {
				s.logger.Info("App container shutdown gracefully")
			}
		}
	})

	return s, nil
}

func initScheduler(s *Server) {
	// 创建任务管理器
	manager := task.NewManager(s.logger, s.sc, s.app)

	// 注册所有任务(业务层控制)
	if err := manager.RegisterTasks(); err != nil {
		s.logger.Error("failed to register tasks", zap.Error(err))
		return
	}

	// 启动任务调度器
	manager.Start()
}

// initLoggerWithConfig 初始化日志器（使用注入的配置）
func initLoggerWithConfig(s *Server, cfg *internalApp.AppConfig) error {
	lg, err := logger.NewLogger(logger.Config{
		Level:      cfg.Log.Level,
		File:       cfg.Log.File,
		Production: cfg.Log.Production,
	})
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	s.logger = lg

	return nil
}

// initValidatorWithLogger 初始化验证器，返回 UniversalTranslator
func initValidatorWithLogger(lg *zap.Logger) (*ut.UniversalTranslator, error) {
	customValidator := validator.NewCustomValidator()
	customValidator.Engine()
	binding.Validator = customValidator

	var uni *ut.UniversalTranslator

	validate, ok := binding.Validator.Engine().(*validatorV10.Validate)
	if ok {

		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		uni = ut.New(en.New(), en.New(), zh.New())

		zhTran, _ := uni.GetTranslator("zh")
		enTran, _ := uni.GetTranslator("en")

		err := zh_translations.RegisterDefaultTranslations(validate, zhTran)
		if err != nil {
			return nil, err
		}
		err = en_translations.RegisterDefaultTranslations(validate, enTran)
		if err != nil {
			return nil, err
		}
	}

	return uni, nil
}

// initDatabaseWithConfig 初始化数据库（使用注入的配置）
func initDatabaseWithConfig(cfg *internalApp.AppConfig, lg *zap.Logger) (*gorm.DB, error) {
	// 转换 AppConfig.DatabaseConfig 为 dao.DatabaseConfig
	dbConfig := dao.DatabaseConfig{
		Type:            cfg.Database.Type,
		Path:            cfg.Database.Path,
		UserName:        cfg.Database.UserName,
		Password:        cfg.Database.Password,
		Host:            cfg.Database.Host,
		Name:            cfg.Database.Name,
		TablePrefix:     cfg.Database.TablePrefix,
		AutoMigrate:     cfg.Database.AutoMigrate,
		Charset:         cfg.Database.Charset,
		ParseTime:       cfg.Database.ParseTime,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		RunMode:         cfg.Server.RunMode,
	}

	db, err := dao.NewDBEngineWithConfig(dbConfig, lg)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// initStorageWithConfig 初始化存储目录（使用注入的配置）
func initStorageWithConfig(cfg *internalApp.AppConfig) error {
	dirs := []string{
		filepath.Dir(cfg.Log.File),
		cfg.App.TempPath,
		cfg.App.UploadSavePath,
		filepath.Dir(cfg.Database.Path),
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0754); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// GetApp 获取 App Container
func (s *Server) GetApp() *internalApp.App {
	return s.app
}

// GetConfig 获取应用配置
func (s *Server) GetConfig() *internalApp.AppConfig {
	return s.config
}
