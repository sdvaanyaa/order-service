package consumer

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/sdvaanyaa/order-service/internal/config"
	"github.com/sdvaanyaa/order-service/internal/models"
	"github.com/sdvaanyaa/order-service/internal/service"
	"log/slog"
	"math"
	"math/rand"
	"time"
)

const (
	MaxAddOrderRetries = 5
	BaseDelay          = 1 * time.Second
	MaxAddOrderDelay   = 30 * time.Second
	MaxConsumeDelay    = 60 * time.Second
	BackoffFactor      = 2
)

type Consumer interface {
	Run(ctx context.Context)
	Ready() <-chan bool
}

type kafkaConsumer struct {
	group   sarama.ConsumerGroup
	handler *Handler
	log     *slog.Logger
	topics  []string
}

type Handler struct {
	svc   service.OrderService
	log   *slog.Logger
	ready chan bool
}

func New(cfg config.KafkaConfig, svc service.OrderService, log *slog.Logger) (Consumer, error) {
	kconf := sarama.NewConfig()

	kconf.Consumer.Offsets.Initial = sarama.OffsetOldest
	kconf.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.Group, kconf)
	if err != nil {
		return nil, err
	}

	return &kafkaConsumer{
		group: group,
		handler: &Handler{
			svc:   svc,
			log:   log,
			ready: make(chan bool),
		},
		log:    log,
		topics: []string{cfg.Topic},
	}, nil
}

func (c *kafkaConsumer) Ready() <-chan bool {
	return c.handler.ready
}

func (c *kafkaConsumer) Run(ctx context.Context) {
	go func() {
		for err := range c.group.Errors() {
			c.log.Error("kafka group error", slog.Any("error", err))
		}
	}()

	for {
		if err := c.group.Consume(ctx, c.topics, c.handler); err != nil {
			c.log.Error("kafka consume failed", slog.Any("error", err))
			delay := backoffDelay(1, BaseDelay, MaxConsumeDelay)
			time.Sleep(delay)
		}
		select {
		case <-ctx.Done():
			c.log.Info("Kafka consumer stopping due to context cancellation")
			if err := c.group.Close(); err != nil {
				c.log.Error("failed to close Kafka consumer group", slog.Any("error", err))
			}
			return
		default:
			c.handler.ready = make(chan bool)
		}
	}
}

func (h *Handler) Setup(sarama.ConsumerGroupSession) error {
	h.log.Info("consumer setup complete, ready to consume")
	close(h.ready)
	return nil
}

func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			h.log.Info(
				"message received",
				slog.String("topic", msg.Topic),
				slog.Int("partition", int(msg.Partition)),
				slog.Int64("offset", msg.Offset),
			)
			h.processMessage(session, msg)
		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *Handler) processMessage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	var order models.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		h.log.Error("unmarshal failed", slog.Any("error", err))
		session.MarkMessage(msg, "")
		return
	}

	h.log.Info("processing order", slog.String("order_uid", order.OrderUID))

	err := h.tryAddOrder(session.Context(), &order)
	if err != nil {
		h.log.Error("add order failed after retries", slog.Any("error", err))
		// consider DLQ in prod
	} else {
		h.log.Info("order processed", slog.String("order_uid", order.OrderUID))
	}

	session.MarkMessage(msg, "")
}

func (h *Handler) tryAddOrder(ctx context.Context, order *models.Order) error {
	attempt := 0
	var addErr error
	for attempt < MaxAddOrderRetries {
		addErr = h.svc.AddOrder(ctx, order)
		if addErr == nil {
			return nil
		}

		attempt++
		h.log.Warn("add order retry", slog.Int("attempt", attempt), slog.Any("error", addErr))
		delay := backoffDelay(attempt, BaseDelay, MaxAddOrderDelay)
		time.Sleep(delay)
	}
	return addErr
}

func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	exp := math.Pow(BackoffFactor, float64(attempt-1))
	delay := time.Duration(exp) * base
	if delay > max {
		delay = max
	}
	jitter := time.Duration(rand.Int63n(int64(delay)))
	return jitter
}
