package api

import (
	"net/http"
	"suscord/internal/domain/entity"
	"suscord/internal/transport/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func (h *handler) InitAuthRoutes(route *echo.Group) {
	route.POST("/auth/login", h.login)
	route.POST("/auth/logout", h.logout, h.middleware.RequiredAuth())
}

type loginInput struct {
	Username string `json:"username" validate:"required,gte=1,lte=15"`
	Password string `json:"password" validate:"required"`
}

func (h *handler) login(c echo.Context) error {
	input := new(loginInput)

	if err := c.Bind(input); err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		return utils.NewErrorResponse(c, http.StatusBadRequest, err)
	}

	sessionID, err := h.service.Auth().Login(c.Request().Context(), &entity.LoginOrCreateInput{
		Username: input.Username,
		Password: input.Password,
	})
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.IsTLS(),
	})

	return c.NoContent(http.StatusOK)
}

func (h *handler) logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "session_id",
		Value:    "",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Secure:   c.IsTLS(),
	})
	return c.NoContent(http.StatusOK)
}
