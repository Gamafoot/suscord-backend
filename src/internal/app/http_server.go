package app

import (
	"context"
	"fmt"
	"io"
	"suscord/internal/config"
	"suscord/internal/domain/eventbus"
	"suscord/internal/domain/storage"
	"suscord/internal/service"
	"suscord/internal/transport/api"
	customMiddleware "suscord/internal/transport/middleware"
	"suscord/internal/transport/ws"
	"suscord/internal/transport/ws/hub"

	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type httpServer struct {
	cfg  *config.Config
	echo *echo.Echo
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewHttpServer(
	cfg *config.Config,
	service service.Service,
	storage storage.Storage,
	eventbus eventbus.EventBus,
	logger *zap.SugaredLogger,
) *httpServer {
	server := &httpServer{
		cfg:  cfg,
		echo: echo.New(),
	}

	// Добавить require_auth для статики
	server.echo.Static(server.cfg.Static.URL, server.cfg.Static.Path)
	server.echo.Static(server.cfg.Media.Url, server.cfg.Media.Path)

	server.echo.Validator = &CustomValidator{validator: validator.New()}

	_customMiddleware := customMiddleware.NewMiddleware(cfg, storage)

	server.echo.Use(
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     server.cfg.Secure.CORS.Origins,
			AllowMethods:     server.cfg.Secure.CORS.AllowedMethods,
			AllowHeaders:     server.cfg.Secure.CORS.AllowedHeaders,
			AllowCredentials: true,
		}),
	)

	apiHandler := api.NewHandler(
		server.cfg,
		service,
		storage,
		_customMiddleware,
	)

	apiRoute := server.echo.Group(
		"/api",
		// middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		// 	Timeout: server.cfg.Server.Timeout,
		// }),
		// middleware.BodyLimit(server.cfg.Media.MaxSize),
		_customMiddleware.AllowedFileExtentions(),
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}),
		customMiddleware.RequestLogger,
	)

	apiHandler.InitRoutes(apiRoute)

	hub := hub.NewHub(
		cfg,
		storage,
		eventbus,
		logger,
	)

	wsRoute := server.echo.Group("", _customMiddleware.RequiredAuth())

	wsHandler := ws.NewHandler(cfg, hub, storage, logger)
	wsHandler.InitRoutes(wsRoute)

	return server
}

func (s *httpServer) Run() error {
	addr := fmt.Sprintf("%s:%s", s.cfg.Api.Host, s.cfg.Api.Port)
	if err := s.echo.Start(addr); err != nil {
		return err
	}

	return nil
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

func (s *httpServer) Echo() *echo.Echo {
	return s.echo
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
