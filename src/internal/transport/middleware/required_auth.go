package middleware

import (
	"net/http"
	domainErrors "suscord/internal/domain/errors"

	"github.com/labstack/echo/v4"
	pkgerr "github.com/pkg/errors"
)

func (mw *Middleware) RequiredAuth() func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session_id")
			if err != nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			ctx := c.Request().Context()

			session, err := mw.storage.Session().GetByUUID(ctx, cookie.Value)
			if err != nil {
				if pkgerr.Is(err, domainErrors.ErrRecordNotFound) {
					return c.NoContent(http.StatusUnauthorized)
				}
				return err
			}

			c.Set("user_id", session.UserID)

			return next(c)
		}
	}
}
