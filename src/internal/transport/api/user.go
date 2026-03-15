package api

import (
	"errors"
	"net/http"
	derr "suscord/internal/domain/errors"
	"suscord/internal/transport/dto"
	"suscord/internal/transport/utils"

	"github.com/labstack/echo/v4"
)

func (h *handler) InitUserRoutes(route *echo.Group) {
	route.GET("/users/:user_id", h.getUserInfo)
	route.GET("/users", h.searchUsers)
}

func (h *handler) getUserInfo(c echo.Context) error {
	userID, err := utils.GetUIntFromParam(c, "user_id")
	if err != nil {
		return utils.NewStringResponse(c, http.StatusBadRequest, "user_id must be digit")
	}

	user, err := h.service.User().GetByID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, derr.ErrRecordNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	return c.JSON(http.StatusOK, dto.NewUser(user, h.mediaURL()))
}

func (h *handler) searchUsers(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	searchPattern := c.QueryParam("search")

	users, err := h.service.User().SearchUsers(c.Request().Context(), userID, searchPattern)
	if err != nil {
		if errors.Is(err, derr.ErrRecordNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	return c.JSON(http.StatusOK, dto.NewUsers(users, h.mediaURL()))
}
