package api

import (
	"errors"
	"net/http"
	derr "suscord/internal/domain/errors"
	"suscord/internal/transport/dto"
	"suscord/internal/transport/utils"

	"github.com/labstack/echo/v4"
	pkgerr "github.com/pkg/errors"
)

func (h *handler) InitChatMemberRoutes(route *echo.Group) {
	route.GET("/chats/:chat_id/members", h.getChatMembers)
	route.GET("/chats/:chat_id/non-members", h.getNonMembers)
	route.POST("/chats/:chat_id/invite", h.sendInviteChat)
	route.GET("/chats/invite/accept/:code", h.acceptInviteChat)
	route.GET("/chats/:chat_id/leave", h.leaveFromChat)
}

func (h *handler) getChatMembers(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	chatID, err := utils.GetUIntFromParam(c, "chat_id")
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	users, err := h.service.ChatMember().GetNonMembers(c.Request().Context(), chatID, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewUsers(users, h.mediaURL()))
}

func (h *handler) getNonMembers(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	chatID, err := utils.GetUIntFromParam(c, "chat_id")
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	users, err := h.service.ChatMember().GetNotChatMembers(c.Request().Context(), chatID, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewUsers(users, h.mediaURL()))
}

func (h *handler) sendInviteChat(c echo.Context) error {
	reqInput := new(dto.InviteUserRequest)

	if err := c.Bind(reqInput); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(reqInput); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	chatID, err := utils.GetUIntFromParam(c, "chat_id")
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	err = h.service.ChatMember().SendInvite(
		c.Request().Context(),
		userID,
		chatID,
		reqInput.UserID,
	)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *handler) acceptInviteChat(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	code := c.Param("code")

	err := h.service.ChatMember().AcceptInvite(c.Request().Context(), userID, code)
	if err != nil {
		if errors.Is(err, derr.ErrKeyNotFound) {
			return c.NoContent(http.StatusGone)
		}
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *handler) leaveFromChat(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	chatID, err := utils.GetUIntFromParam(c, "chat_id")
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	err = h.service.ChatMember().LeaveFromChat(c.Request().Context(), chatID, userID)
	if err != nil {
		if errors.Is(err, derr.ErrUserIsNotMemberOfChat) {
			return utils.NewErrorResponse(c, http.StatusNotFound, err)
		}
		return err
	}

	return c.NoContent(http.StatusOK)
}
