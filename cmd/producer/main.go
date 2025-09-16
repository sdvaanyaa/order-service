package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/sdvaanyaa/order-service/internal/config"
	"github.com/sdvaanyaa/order-service/internal/models"
	"log/slog"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("config load failed", "err", err)
		os.Exit(1)
	}

	pconf := sarama.NewConfig()
	pconf.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, pconf)
	if err != nil {
		log.Error("producer init failed", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err = producer.Close(); err != nil {
			log.Error("producer close failed", "err", err)
		}
	}()

	for i := 0; i < 5; i++ {
		order := generateRandomOrder()
		msgBytes, err := json.Marshal(order)
		if err != nil {
			log.Error("marshal failed", "err", err)
			continue
		}
		msg := &sarama.ProducerMessage{
			Topic: cfg.Kafka.Topic,
			Value: sarama.ByteEncoder(msgBytes),
		}
		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Error("send failed", "err", err)
			continue
		}
		log.Info(
			"order sent",
			slog.String("uid", order.OrderUID),
			slog.Int("partition", int(partition)),
			slog.Int64("offset", offset),
		)

		time.Sleep(2 * time.Second)
	}
}

func generateRandomOrder() models.Order {
	uid := uuid.NewString()
	return models.Order{
		OrderUID:    uid,
		TrackNumber: fmt.Sprintf("TRACK%d", rand.Intn(10000)),
		Entry:       "ENTRY",
		Delivery: models.Delivery{
			Name:    "Test User",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "City",
			Address: "Address 1",
			Region:  "Region",
			Email:   "test@email.com",
		},
		Payment: models.Payment{
			Transaction:  uid,
			Currency:     "USD",
			Provider:     "pay",
			Amount:       rand.Intn(1000) + 100,
			PaymentDt:    time.Now().Unix(),
			Bank:         "bank",
			DeliveryCost: 50,
			GoodsTotal:   100,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      int64(rand.Intn(100000)),
				TrackNumber: fmt.Sprintf("ITEM%d", rand.Intn(10000)),
				Price:       rand.Intn(500) + 50,
				Rid:         "rid",
				Name:        "Item Name",
				Sale:        rand.Intn(50),
				Size:        "M",
				TotalPrice:  100,
				NmID:        int64(rand.Intn(100000)),
				Brand:       "Brand",
				Status:      200,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "cust",
		DeliveryService:   "service",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}
