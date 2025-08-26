package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sdvaanyaa/order-service.git/internal/config"
	"github.com/sdvaanyaa/order-service.git/internal/handler"
	"github.com/sdvaanyaa/order-service.git/internal/repository/postgres"
	"github.com/sdvaanyaa/order-service.git/internal/service"
	"github.com/sdvaanyaa/order-service.git/pkg/pgdb"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("config load failed", "err", err)
		os.Exit(1)
	}

	db, err := pgdb.New(cfg.Postgres, log)
	if err != nil {
		log.Error("db init failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	transactor := pgdb.NewTransactor(db)
	repo := postgres.New(db, log)
	svc := service.New(repo, transactor, log)
	h := handler.New(svc)

	app := fiber.New()
	h.SetupRoutes(app, log)

	go func() {
		if err = app.Listen(cfg.HTTP.Address()); err != nil {
			log.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	if err = app.Shutdown(); err != nil {
		log.Error("shutdown failed", slog.Any("error", err))
	}
}
