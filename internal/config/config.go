package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log/slog"
)

type Config struct {
	Postgres PostgresConfig
	HTTP     HTTPConfig
	Kafka    KafkaConfig
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	Username string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	Database string `env:"POSTGRES_DB" envDefault:"orders"`
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
	)
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT" envDefault:"8080"`
}

func (c HTTPConfig) Address() string {
	return ":" + c.Port
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	Topic   string   `env:"KAFKA_TOPIC" envDefault:"orders"`
	Group   string   `env:"KAFKA_GROUP" envDefault:"order-service"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found", "err", err)
	}

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
