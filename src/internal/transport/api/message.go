package api

import (
	"errors"
	"fmt"
	"net/http"
	"suscord/internal/domain/entity"
	derr "suscord/internal/domain/errors"
	domainerr "suscord/internal/domain/errors"
	"suscord/internal/transport/dto"
	"suscord/internal/transport/utils"

	"github.com/labstack/echo/v4"
	pkgerr "github.com/pkg/errors"
)

func (h *handler) InitMessageRoutes(route *echo.Group) {
	route.GET("/chats/:chat_id/messages", h.getChatMessages)
	route.POST("/chats/:chat_id/messages", h.createMessage)
	route.PATCH("/messages/:message_id", h.updateMessage)
	route.DELETE("/messages/:message_id", h.deleteMessage)
}

func (h *handler) getChatMessages(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	chatID, err := utils.GetUIntFromParam(c, "chat_id")
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	lastMessageID, err := utils.GetIntFromQuery(c, "last_message_id", 0)
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	limit, err := utils.GetIntFromQuery(c, "limit", 10)
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	messages, err := h.service.Message().GetChatMessages(
		c.Request().Context(),
		entity.GetMessagesInput{
			ChatID:        chatID,
			UserID:        userID,
			LastMessageID: uint(lastMessageID),
			Limit:         limit,
		},
	)
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, dto.NewMessages(messages, h.mediaURL()))
}

type createMessageRequest struct {
	Type    string `form:"type" validate:"min=1,max=10"`
	Content string `form:"content"`
}

func (h *handler) createMessage(c echo.Context) error {
	req := new(createMessageRequest)

	if err := c.Bind(req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(req); err != nil {
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

	form, err := c.MultipartForm()
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return pkgerr.WithStack(err)
	}

	var eFiles []*entity.File

	defer func() {
		for _, file := range eFiles {
			if file.Reader != nil {
				file.Reader.Close()
			}
		}
	}()

	if form != nil && form.File != nil {
		files := form.File["file"]

		if len(files) > 5 {
			return utils.NewStringResponse(c, http.StatusBadRequest, fmt.Sprintf("too many files: %d", len(files)))
		}

		eFiles = make([]*entity.File, 0, len(files))

		for _, file := range files {
			mimetype := utils.GetMimeType(file.Filename)

			if !utils.FilenameValidate(file.Filename, h.cfg.Media.AllowedMedia) {
				text := fmt.Sprintf("invalid file media (%s): %s", mimetype, file.Filename)
				return utils.NewStringResponse(c, http.StatusBadRequest, text)
			}

			src, err := file.Open()
			if err != nil {
				return pkgerr.WithStack(err)
			}

			eFiles = append(eFiles, &entity.File{
				Name:     file.Filename,
				Size:     file.Size,
				MimeType: mimetype,
				Reader:   src,
			})
		}
	}

	input := entity.CreateMessageInput{
		Type:    req.Type,
		Content: req.Content,
		Files:   eFiles,
	}

	message, err := h.service.Message().Create(
		c.Request().Context(),
		userID,
		chatID,
		input,
	)
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, dto.NewMessage(message, h.mediaURL()))
}

type updateMessageRequest struct {
	Content string `json:"content"`
}

func (h *handler) updateMessage(c echo.Context) error {
	reqInput := new(updateMessageRequest)

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

	messageID, err := utils.GetUIntFromParam(c, "message_id")
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	input := entity.UpdateMessageInput{
		Content: reqInput.Content,
	}

	message, err := h.service.Message().Update(
		c.Request().Context(),
		userID,
		messageID,
		input,
	)
	if err != nil {
		if errors.Is(err, domainerr.ErrRecordNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	return c.JSON(http.StatusOK, dto.NewMessage(message, h.mediaURL()))
}

func (h *handler) deleteMessage(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	messageID, err := utils.GetUIntFromParam(c, "message_id")
	if err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	err = h.service.Message().Delete(c.Request().Context(), userID, messageID)
	if err != nil {
		if errors.Is(err, domainerr.ErrRecordNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	return c.NoContent(http.StatusOK)
}
