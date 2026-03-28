package ws

import (
	"net/http"
	"suscord/internal/config"
	"suscord/internal/domain/storage"
	"suscord/internal/transport/dto"
	"suscord/internal/transport/utils"
	"suscord/internal/transport/ws/hub"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	pkgerr "github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type handler struct {
	cfg      *config.Config
	upgrader *websocket.Upgrader
	hub      hub.Hub
	storage  storage.Storage
	logger   *zap.SugaredLogger
}

func NewHandler(
	config *config.Config,
	hub hub.Hub,
	storage storage.Storage,
	logger *zap.SugaredLogger,
) *handler {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return lo.Contains(config.Secure.CORS.Origins, origin)
		},
	}
	return &handler{
		cfg:      config,
		upgrader: upgrader,
		hub:      hub,
		storage:  storage,
		logger:   logger,
	}
}

func (h *handler) InitRoutes(route *echo.Group) {
	route.GET("/ws", h.websocket)
	route.GET("/api/current-call-room/members", h.getCurrentCallMembers)
}

var (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second
)

func (h *handler) websocket(c echo.Context) error {
	cookie, err := c.Cookie("session_id")
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	if cookie == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	conn, err := h.upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return pkgerr.WithStack(err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	userID := c.Get("user_id").(uint)

	user, err := h.storage.User().GetByID(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	chats, err := h.storage.Chat().GetUserChats(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	client := hub.NewHubClient(conn, user)

	h.hub.Register(client, chats)
	h.hub.ReceiveMessageHandler(client)

	h.hub.Unregister(client)

	return nil
}

func (h *handler) getCurrentCallMembers(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	members, err := h.hub.GetCurrentCallMembers(userID)
	if err != nil {
		return utils.NewStringResponse(c, http.StatusNotFound, err.Error())
	}

	result := make([]dto.User, len(members))
	for i, member := range members {
		result[i] = dto.NewUser(member, h.cfg.Media.Url)
	}

	return c.JSON(http.StatusOK, result)
}
