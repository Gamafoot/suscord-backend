package api

import (
	"fmt"
	"net/http"
	derr "suscord/internal/domain/errors"
	"suscord/internal/transport/utils"

	"github.com/labstack/echo/v4"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	pkgerr "github.com/pkg/errors"
)

func (h *handler) InitLivekitRoutes(route *echo.Group) {
	route.GET("/call/get_token", h.getToken)
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (h *handler) getToken(c echo.Context) error {
	room := c.QueryParam("room")

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return derr.ErrNoContextVar
	}

	if room == "" {
		return utils.NewStringResponse(c, http.StatusBadRequest, "room and identity are required")
	}

	if h.cfg.LiveKit.ApiKey == "" || h.cfg.LiveKit.ApiSecret == "" {
		return utils.NewStringResponse(c, http.StatusInternalServerError, "livekit api key/secret not configured")
	}

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	grant.SetCanPublish(true)
	grant.SetCanPublishSources([]livekit.TrackSource{
		livekit.TrackSource_MICROPHONE,
		livekit.TrackSource_SCREEN_SHARE,
		livekit.TrackSource_SCREEN_SHARE_AUDIO,
	})

	at := auth.NewAccessToken(h.cfg.LiveKit.ApiKey, h.cfg.LiveKit.ApiSecret)
	at = at.AddGrant(grant)
	at = at.SetIdentity(fmt.Sprint(userID))

	token, err := at.ToJWT()
	if err != nil {
		return pkgerr.WithStack(err)
	}

	return c.JSON(http.StatusOK, tokenResponse{Token: token})
}
