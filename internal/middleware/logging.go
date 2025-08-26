package middleware

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"time"
)

func Logging(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		log.Info("HTTP request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("ip", c.IP()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", duration),
		)

		return err
	}
}
