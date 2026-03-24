package app

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/workerpool"
	"github.com/haierkeys/fast-note-sync-service/pkg/writequeue"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Infra encapsulates infrastructure dependencies
type Infra struct {
	config         *AppConfig
	logger         *zap.Logger
	DB             *gorm.DB
	Dao            *dao.Dao
	workerPool     *workerpool.Pool
	writeQueueMgr  *writequeue.Manager
	TokenManager   pkgapp.TokenManager
	sourceSelector *fileurl.SourceSelector
}

// initInfra initializes infrastructure components
func initInfra(cfg *AppConfig, logger *zap.Logger, db *gorm.DB) (*Infra, error) {
	infra := &Infra{
		config:         cfg,
		logger:         logger,
		DB:             db,
		sourceSelector: fileurl.NewSourceSelector(cfg.App.PullSource),
	}

	// Worker Pool
	wpConfig := cfg.GetWorkerPoolConfig()
	infra.workerPool = workerpool.New(&wpConfig, logger)

	// Write Queue Manager
	wqConfig := cfg.GetWriteQueueConfig()
	infra.writeQueueMgr = writequeue.New(&wqConfig, logger)

	// DAO
	dbConfig := &dao.DatabaseConfig{
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
	infra.Dao = dao.New(db, context.Background(),
		dao.WithConfig(dbConfig),
		dao.WithLogger(logger),
		dao.WithWriteQueueManager(infra.writeQueueMgr),
	)

	// TokenManager
	tokenConfig := pkgapp.TokenConfig{
		SecretKey:     cfg.Security.AuthTokenKey,
		Issuer:        "fast-note-sync-service",
		Expiry:        cfg.GetTokenExpiry(),
		ShareTokenKey: cfg.Security.ShareTokenKey,
		ShareExpiry:   cfg.GetShareTokenExpiry(),
	}
	infra.TokenManager = pkgapp.NewTokenManager(tokenConfig)

	return infra, nil
}
