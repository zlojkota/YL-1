package main

import (
	"context"
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
	"github.com/zlojkota/YL-1/internal/serverhelpers"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	ServerAddr    *string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval *time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     *string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       *bool          `env:"RESTORE" envDefault:"true"`
}

var cfg Config

func init() {

}

func main() {
	// Setup

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if *cfg.ServerAddr == "127.0.0.1:8080" {
		cfg.ServerAddr = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	}
	if *cfg.StoreFile == "/tmp/devops-metrics-db.json" {
		cfg.StoreFile = flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE")
	}
	if *cfg.Restore {
		cfg.Restore = flag.Bool("r", true, "RESTORE")
	}
	if *cfg.StoreInterval == 300*time.Second {
		cfg.StoreInterval = flag.Duration("i", 300*time.Second, "STORE_INTERVAL")
	}
	flag.Parse()

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

	var helper serverhelpers.StorageState
	helper.SetServerHandler(handler)

	if *cfg.Restore {
		helper.Restore(*cfg.StoreFile)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan
		log.Error("Stopping")
		helper.Done <- true
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	go helper.Run(*cfg.StoreInterval, *cfg.StoreFile)
	if err := e.Start(*cfg.ServerAddr); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal("shutting down the server")
	}

}
