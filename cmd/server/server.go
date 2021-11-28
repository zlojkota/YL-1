package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/serverHeaders"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Setup
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	//default answer
	e.GET("/*", serverHeaders.DefaultHandler)
	e.POST("/*", serverHeaders.DefaultHandler)
	// update Handler
	e.POST("/:method/:type/:metric/:value", serverHeaders.UpdateHandler)
	e.POST("/update/*", serverHeaders.UpdateHandler)

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
