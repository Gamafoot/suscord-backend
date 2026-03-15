package api

import (
	"suscord/internal/config"
	"suscord/internal/domain/storage"
	"suscord/internal/service"
	"suscord/internal/transport/middleware"

	"github.com/labstack/echo/v4"
)

type handler struct {
	cfg        *config.Config
	service    service.Service
	storage    storage.Storage
	middleware *middleware.Middleware
}

func NewHandler(
	config *config.Config,
	service service.Service,
	storage storage.Storage,
	middleware *middleware.Middleware,
) *handler {
	return &handler{
		cfg:        config,
		service:    service,
		storage:    storage,
		middleware: middleware,
	}
}

func (h *handler) InitRoutes(route *echo.Group) {
	h.InitAuthRoutes(route)

	requiredAuth := route.Group("", h.middleware.RequiredAuth())
	h.InitMeRoutes(requiredAuth)
	h.InitUserRoutes(requiredAuth)
	h.InitChatRoutes(requiredAuth)
	h.InitChatMemberRoutes(requiredAuth)
	h.InitMessageRoutes(requiredAuth)
	h.InitLivekitRoutes(requiredAuth)
}

func (h *handler) mediaURL() string {
	return h.cfg.Media.Url
}
