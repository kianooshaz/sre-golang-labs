package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type HelloWorldResponse struct {
	Message string `json:"message"`
}

func HelloWorld(c *echo.Context) error {
	resp := HelloWorldResponse{
		Message: "Hello, World!",
	}

	return c.JSON(http.StatusOK, resp)
}
