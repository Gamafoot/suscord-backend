package app

import (
	"context"
	"log"
	"suscord/internal/config"
	"suscord/internal/infra/cache/inmemory"
	"suscord/internal/infra/database/relational"
	implStorage "suscord/internal/infra/database/relational/storage"
	"suscord/internal/infra/eventbus"
	file "suscord/internal/infra/file_manager"
	"suscord/internal/service"
	"suscord/pkg/logger"

	"github.com/pkg/errors"
)

type App struct {
	httpServer *httpServer
	shutdown   func()
}

func NewApp() (*App, error) {
	cfg := config.GetConfig()

	db, err := relational.NewConnect(cfg.Database.URL, cfg.Database.LogLevel)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	storage := implStorage.NewGormStorage(db)

	logger, cleanup, err := logger.NewSugaredLogger(logger.Config{
		FilePath:   "logs/app.log",
		MaxSizeMB:  100,
		MaxBackups: 5,
		MaxAgeDays: 30,
		Compress:   true,
		Level:      cfg.Logger.Level,
	})
	if err != nil {
		panic(err)
	}

	cache := inmemory.NewCache()
	eventbus := eventbus.NewEventBus(logger, cfg.EventBus.Timeout)
	fileManager := file.NewFileManager(cfg.Media.Path)

	service := service.NewService(service.ServiceConfig{
		AppConfig:   cfg,
		Storage:     storage,
		Cache:       cache,
		Eventbus:    eventbus,
		FileManager: fileManager,
		Logger:      logger,
	})

	app := new(App)
	app.httpServer = NewHttpServer(cfg, service, storage, eventbus, logger)
	app.shutdown = func() {
		cleanup()

		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("fail get sql.DB: %v", err)
		}

		if err = sqlDB.Close(); err != nil {
			log.Printf("fail close database connection: %v", err)
		}
	}

	return app, nil
}

func (a *App) RunApi() error {
	return a.httpServer.Run()
}

func (a *App) RunWebsocket(ctx context.Context) error {
	return a.httpServer.Run()
}

func (a *App) Shutdown(ctx context.Context) {
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Printf("fail server shutdown: %v", err)
	}
	a.shutdown()
}
