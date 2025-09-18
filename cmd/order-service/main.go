package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sdvaanyaa/order-service/internal/config"
	"github.com/sdvaanyaa/order-service/internal/consumer"
	"github.com/sdvaanyaa/order-service/internal/handler"
	"github.com/sdvaanyaa/order-service/internal/repository/postgres"
	"github.com/sdvaanyaa/order-service/internal/service"
	"github.com/sdvaanyaa/order-service/pkg/pgdb"
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
	val := validator.New()
	svc := service.New(repo, transactor, log, val)
	h := handler.New(svc)

	cons, err := consumer.New(cfg.Kafka, svc, log)
	if err != nil {
		log.Error("kafka consumer init failed", "err", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cons.Run(ctx)
	<-cons.Ready()
	log.Info("kafka consumer ready")

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
