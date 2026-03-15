package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func NewStringResponse(c echo.Context, status int, message string) error {
	return c.String(status, fmt.Sprintf(`{"error": "%s"}`, message))
}

func NewErrorResponse(c echo.Context, status int, err error) error {
	if err == nil {
		return nil
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fe := range ve {
			errors[fe.Field()] = fmt.Sprintf("rule: %s, limit: %s", fe.Tag(), fe.Param())
		}
		return c.JSON(status, map[string]any{"errors": errors})
	}

	return c.JSON(status, map[string]string{"error": err.Error()})
}
