package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type AddServerRequest struct {
	Name string `json:"name" validate:"required"`
	IP   string `json:"ip" validate:"required"`
}

type AddServerResponse struct {
	Message string `json:"message"`
}

func AddServer(c *echo.Context) error {
	var req AddServerRequest
	var res AddServerResponse

	err := c.Bind(&req)
	if err != nil {
		res.Message = "Invalid request payload"
		return c.JSON(http.StatusBadRequest, res)
	}

	// use validator
	if err := c.Validate(&req); err != nil {
		res.Message = "Validation failed: " + err.Error()
		return c.JSON(http.StatusBadRequest, res)
	}

	// insert db logic here to add the server to the database

	res.Message = "Server added successfully"

	return c.JSON(http.StatusOK, res)
}
