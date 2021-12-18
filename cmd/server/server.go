package main

import (
	"context"
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/dbstorage"
	"github.com/zlojkota/YL-1/internal/filestorage"
	"github.com/zlojkota/YL-1/internal/serverhandlers"
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
	HashKey       *string        `env:"KEY" envDefault:""`
	DatabaseDsn   *string        `env:"DATABASE_DSN" envDefault:""`
}

type StorageHelper interface {
	Run(storeInterval time.Duration)
	Restore()
	SendDone()
	WaitDone()
	Init(serverHandler *serverhandlers.ServerHandler, store string)
	Ping() bool
}

func main() {
	// Setup
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		cfg.ServerAddr = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	} else {
		_ = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	}
	if _, ok := os.LookupEnv("STORE_FILE"); !ok {
		cfg.StoreFile = flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE")
	} else {
		_ = flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE")
	}
	if _, ok := os.LookupEnv("RESTORE"); !ok {
		cfg.Restore = flag.Bool("r", true, "RESTORE")
	} else {
		_ = flag.Bool("r", true, "RESTORE")
	}
	if _, ok := os.LookupEnv("STORE_INTERVAL"); !ok {
		cfg.StoreInterval = flag.Duration("i", 300*time.Second, "STORE_INTERVAL")
	} else {
		_ = flag.Duration("i", 300*time.Second, "STORE_INTERVAL")
	}
	if _, ok := os.LookupEnv("KEY"); !ok {
		cfg.HashKey = flag.String("k", "", "KEY")
	} else {
		_ = flag.String("k", "", "KEY")
	}
	if _, ok := os.LookupEnv("DATABASE_DSN"); !ok {
		cfg.DatabaseDsn = flag.String("d", "", "DATABASE_DSN")
	} else {
		_ = flag.String("d", "", "DATABASE_DSN")
	}
	flag.Parse()

	e := echo.New()
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} method=${method}, uri=${uri}, status=${status} Content-Type=${header:Content-Type} =${header:Content-Type}\n",
	}))

	handler := serverhandlers.NewServerHandler()
	handler.SetHasher(*cfg.HashKey)

	var helper StorageHelper

	if *cfg.DatabaseDsn != "" {
		helper = new(dbstorage.DataBaseStorageState)
		helper.Init(handler, *cfg.DatabaseDsn)

	} else {
		helper = new(filestorage.FileStorageState)
		helper.Init(handler, *cfg.StoreFile)
	}
	//default answer
	e.GET("/*", handler.NotFoundHandler)
	e.POST("/*", handler.NotFoundHandler)

	// update Handler
	e.POST("/update/:type/:metric/:value", handler.UpdateHandler)
	e.POST("/update/", handler.UpdateHandler)

	// homePage Handler
	e.GET("/", handler.MainHandler)

	// getValue Handler
	e.GET("/value/:type/:metric", handler.GetHandler)
	e.POST("/value/", handler.GetHandler)

	//ping DB
	e.GET("/ping", func(c echo.Context) error {
		if helper.Ping() {
			return c.NoContent(http.StatusOK)
		}
		return c.NoContent(http.StatusInternalServerError)
	})

	if *cfg.Restore {
		helper.Restore()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan
		helper.SendDone()
		log.Error("Stopping")
		helper.WaitDone()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	go helper.Run(*cfg.StoreInterval)
	if err := e.Start(*cfg.ServerAddr); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal("shutting down the server")
	}

}
