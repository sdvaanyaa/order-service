package handler

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/sdvaanyaa/order-service/internal/middleware"
	"github.com/sdvaanyaa/order-service/internal/models"
	"github.com/sdvaanyaa/order-service/internal/repository"
	"github.com/sdvaanyaa/order-service/internal/service"
	"log/slog"
)

type Handler struct {
	svc service.OrderService
}

func New(svc service.OrderService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SetupRoutes(app *fiber.App, log *slog.Logger) {
	app.Use(middleware.Logging(log))
	app.Post("/order", h.AddOrder)
	app.Get("/order/:uid", h.GetOrder)
	app.Get("/", h.Index)
}

func (h *Handler) AddOrder(c *fiber.Ctx) error {
	var order models.Order

	if err := json.Unmarshal(c.Body(), &order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}
	if err := h.svc.AddOrder(c.Context(), &order); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrOrderAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) GetOrder(c *fiber.Ctx) error {
	uid := c.Params("uid")

	order, err := h.svc.GetOrder(c.Context(), uid)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	return c.JSON(order)
}

func (h *Handler) Index(c *fiber.Ctx) error {
	return c.SendFile("./static/index.html")
}
