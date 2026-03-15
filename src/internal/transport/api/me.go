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

func (h *handler) InitMeRoutes(route *echo.Group) {
	route.GET("/users/me", h.aboutMe)
	route.PATCH("/users/me", h.updateMe)
}

func (h *handler) aboutMe(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	user, err := h.service.User().GetByID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, derr.ErrRecordNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	return c.JSON(http.StatusOK, dto.NewMe(user, h.mediaURL()))
}

type updateUserRequest struct {
	Username *string               `form:"username" validate:"required_without=File,min=1,max=20"`
	File     *multipart.FileHeader `validate:"required_without=Username"`
}

func (h *handler) updateMe(c echo.Context) error {
	req := updateUserRequest{}

	if err := c.Bind(&req); err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	file, _ := c.FormFile("file")
	if file != nil {
		req.File = file
	}

	if err := c.Validate(&req); err != nil {
		return utils.NewErrorResponse(c, http.StatusUnprocessableEntity, err)
	}

	var (
		eFile *entity.File
		err   error
	)

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

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return pkgerr.WithStack(derr.ErrNoContextVar)
	}

	input := entity.UpdateUserInput{
		Username: req.Username,
		File:     eFile,
	}

	user, err := h.service.User().Update(c.Request().Context(), userID, input)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewMe(user, h.mediaURL()))
}
