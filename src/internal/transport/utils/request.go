package utils

import (
	"strconv"
	domainErrors "suscord/internal/domain/errors"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func GetUIntFromParam(c echo.Context, param string) (uint, error) {
	value := c.Param(param)

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.Wrap(domainErrors.ErrIsNotDigit, param)
	}

	if valueInt < 0 {
		return 0, errors.Wrap(domainErrors.ErrIsNotPositiveDigit, param)
	}

	return uint(valueInt), nil
}

func GetIntFromQuery(c echo.Context, param string, defaultValue ...int) (int, error) {
	value := c.QueryParam(param)

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, errors.Wrap(domainErrors.ErrIsNotDigit, param)
	}

	if valueInt < 0 {
		return 0, errors.Wrap(domainErrors.ErrIsNotPositiveDigit, param)
	}

	return valueInt, nil
}
