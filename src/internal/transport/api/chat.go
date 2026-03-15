package api

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"suscord/internal/domain/entity"
	derr "suscord/internal/domain/errors"
	"suscord/internal/transport/dto"
	"suscord/internal/transport/utils"

	"github.com/labstack/echo/v4"
	pkgerr "github.com/pkg/errors"
)

func (h *handler) InitChatRoutes(route *echo.Group) {
	route.GET("/chats", h.getUserChats)
	route.POST("/chats/private", h.getOrCreatePrivateChat)
	route.POST("/chats/group", h.createGroupChat)
	route.PATCH("/chats/:chat_id", h.updateGroupChat)
	route.DELETE("/chats/:chat_id", h.deletePrivateChat)
}

func (h *handler) getUserChats(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	searchParam := c.QueryParam("search")

	chats, err := h.service.Chat().GetUserChats(c.Request().Context(), userID, searchParam)
	if err != nil {
		return utils.NewStringResponse(c, http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.NewChats(chats, h.mediaURL()))
}

type createPrivateChatRequest struct {
	FriendID uint `json:"friend_id" validate:"required"`
}

func (h *handler) getOrCreatePrivateChat(c echo.Context) error {
	input := new(createPrivateChatRequest)

	if err := c.Bind(input); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(input); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	chat, err := h.service.Chat().GetOrCreatePrivateChat(
		c.Request().Context(),
		userID,
		input.FriendID,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewChat(chat, h.mediaURL()))
}

type createGroupChatRequest struct {
	Name string `json:"name" validate:"required"`
}

func (h *handler) createGroupChat(c echo.Context) error {
	input := new(createGroupChatRequest)

	if err := c.Bind(input); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(input); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	data := entity.CreateGroupChatInput{Name: input.Name}

	chat, err := h.service.Chat().CreateGroupChat(c.Request().Context(), userID, data)
	if err != nil {
		return err
	}

	return c.JSON(
		http.StatusCreated,
		dto.NewChat(chat, h.mediaURL()),
	)
}

type updateGroupChatRequest struct {
	Name *string               `form:"name" validate:"required_without=File,min=1,max=20"`
	File *multipart.FileHeader `validate:"required_without=Name"`
}

func (h *handler) updateGroupChat(c echo.Context) error {
	req := new(updateGroupChatRequest)

	if err := c.Bind(req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	file, _ := c.FormFile("file")
	if file != nil {
		req.File = file
	}

	if err := c.Validate(req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	chatID, err := utils.GetUIntFromParam(c, "chat_id")
	if err != nil {
		return utils.NewStringResponse(c, http.StatusBadRequest, err.Error())
	}

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	isMember, err := h.service.ChatMember().IsMemberOfChat(c.Request().Context(), userID, chatID)
	if err != nil {
		return err
	}

	if !isMember {
		return utils.NewStringResponse(c, http.StatusForbidden, derr.ErrUserIsNotMemberOfChat.Error())
	}

	var eFile *entity.File
	if req.File != nil {
		mimetype := utils.GetMimeType(file.Filename)

		if !utils.IsImage(file.Filename) {
			text := fmt.Sprintf("the file must be image (%s): %s", mimetype, file.Filename)
			return utils.NewStringResponse(c, http.StatusBadRequest, text)
		}

		eFile, err = entity.NewFile(file)
		if err != nil {
			return err
		}
		defer eFile.Close()
	}

	input := entity.UpdateChatInput{
		Name: req.Name,
		File: eFile,
	}

	chat, err := h.service.Chat().UpdateGroupChat(
		c.Request().Context(),
		userID,
		chatID,
		input,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewChat(chat, h.mediaURL()))
}

func (h *handler) deletePrivateChat(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	chatID, err := utils.GetUIntFromParam(c, "chat_id")
	if err != nil {
		return utils.NewStringResponse(c, http.StatusBadRequest, err.Error())
	}

	err = h.service.Chat().DeletePrivateChat(c.Request().Context(), userID, chatID)
	if err != nil {
		if errors.Is(err, derr.ErrForbidden) {
			return utils.NewStringResponse(c, http.StatusForbidden, err.Error())
		} else if errors.Is(err, derr.ErrRecordNotFound) {
			return utils.NewStringResponse(c, http.StatusNotFound, err.Error())
		}
		return err
	}

	return c.NoContent(http.StatusOK)
}
