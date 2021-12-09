package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Setup
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} method=${method}, uri=${uri}, status=${status} Content-Type=${header:Content-Type}\n",
	}))
	handler := serverhandlers.NewServerHandler()

	//default answer
	e.GET("/*", handler.NotFoundHandler)
	e.POST("/*", handler.NotFoundHandler)

	// update Handler
	e.POST("/update/:type/:metric/:value", handler.UpdateHandler)
	e.POST("/update/", handler.UpdateJSONHandler)

	// homePage Handler
	e.GET("/", handler.MainHandler)

	// getValue Handler
	e.GET("/value/:type/:metric", handler.GetHandler)
	e.POST("/value/", handler.GetJSONHandler)

	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
