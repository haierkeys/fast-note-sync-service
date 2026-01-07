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
)

type Server struct {
	logger            *zap.Logger // non-nil logger.
	httpServer        *http.Server
	privateHttpServer *http.Server
	sc                *safe_close.SafeClose
}

func NewServer(runEnv *runFlags) (*Server, error) {

	configRealpath, err := global.ConfigLoad(runEnv.config)
	if err != nil {
		return nil, err
	}

	runMode := runEnv.runMode
	if len(runMode) <= 0 {
		runMode = global.Config.Server.RunMode
	}

	if len(runMode) > 0 {
		gin.SetMode(runMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		sc: safe_close.NewSafeClose(),
	}

	// Init logger.
	initLogger(s)

	// 初始化存储目录
	if err := initStorage(); err != nil {
		return nil, fmt.Errorf("initStorage: %w", err)
	}

	if err := initDatabase(); err != nil {
		return nil, fmt.Errorf("initDatabase: %w", err)
	}

	// 自动执行迁移任务
	if err := upgrade.Execute(); err != nil {
		return nil, fmt.Errorf("upgrade.Execute: %w", err)
	}

	initValidator()

	validator.RegisterCustom()

	// Start scheduler
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

	// Start http api server
	if httpAddr := global.Config.Server.HttpPort; len(httpAddr) > 0 {
		s.logger.Warn("api_router", zap.String("config.server.HttpPort", global.Config.Server.HttpPort))
		s.httpServer = &http.Server{
			Addr:           global.Config.Server.HttpPort,
			Handler:        routers.NewRouter(frontendFiles),
			ReadTimeout:    time.Duration(global.Config.Server.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(global.Config.Server.WriteTimeout) * time.Second,
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

	if httpAddr := global.Config.Server.PrivateHttpListen; len(httpAddr) > 0 {

		s.logger.Info("api_router", zap.String("config.server.PrivateHttpListen", global.Config.Server.PrivateHttpListen))
		s.privateHttpServer = &http.Server{
			Addr:           global.Config.Server.PrivateHttpListen,
			Handler:        routers.NewPrivateRouter(),
			ReadTimeout:    time.Duration(global.Config.Server.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(global.Config.Server.WriteTimeout) * time.Second,
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

	return s, nil
}

func initScheduler(s *Server) {
	// 创建任务管理器
	manager := task.NewManager(s.logger, s.sc)

	// 注册所有任务(业务层控制)
	if err := manager.RegisterTasks(); err != nil {
		s.logger.Error("failed to register tasks", zap.Error(err))
		return
	}

	// 启动任务调度器
	manager.Start()
}

func initLogger(s *Server) error {

	lg, err := logger.NewLogger(logger.Config{Level: global.Config.Log.Level, File: global.Config.Log.File, Production: global.Config.Log.Production})
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	global.Logger = lg
	s.logger = lg

	return nil
}

func initValidator() error {
	global.Validator = validator.NewCustomValidator()
	global.Validator.Engine()
	binding.Validator = global.Validator

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
			return err
		}
		err = en_translations.RegisterDefaultTranslations(validate, enTran)
		if err != nil {
			return err
		}
	}

	global.Ut = uni

	return nil
}

func initDatabase() error {
	var err error
	global.DBEngine, err = dao.NewDBEngine(global.Config.Database)
	if err != nil {
		return err
	}
	return nil
}

// initStorage 初始化存储目录，确保所有必需的目录都已存在。
func initStorage() error {
	dirs := []string{
		filepath.Dir(global.Config.Log.File),
		global.Config.App.TempPath,
		global.Config.App.UploadSavePath,
		filepath.Dir(global.Config.Database.Path),
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
