package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

func StartServer() {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	registerRoutes(e)

	if err := e.Start(":8080"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
