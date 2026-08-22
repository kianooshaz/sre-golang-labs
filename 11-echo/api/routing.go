package api

import (
	"learn-echo/api/handlers"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func registerRoutes(e *echo.Echo) {
	gv1 := e.Group("/api/v1", middleware.Gzip())
	gv1.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://google.com",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"OPTIONS",
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
		},
	}))

	gv1.GET("/hello-world", handlers.HelloWorld)
	gv1.POST("/servers", handlers.AddServer, middleware.BodyLimit(2_097_152))
}
