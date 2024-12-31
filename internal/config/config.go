package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultIsProduction           = "false"
	defaultPostgresUrl            = "postgresql://postgres:postgres@localhost:5432/queue_db"
	defaultRabbitMqUrl            = "amqp://guest:guest@localhost:5672"
	defaultSubscriberExchangeName = "schedule.topic"
	defaultSubscriberBindingKey   = "schedule.parsed"
	defaultSubscriberQueueName    = "queue.service"
	defaultPublisherExchangeName  = "queue.topic"
	defaultPublisherRoutingKey    = "queue.updates.full"
)

type Config struct {
	IsProduction           bool
	PostgresUrl            string
	RabbitMqUrl            string
	PublisherRoutingKey    string
	PublisherExchangeName  string
	SubscriberBindingKey   string
	SubscriberExchangeName string
	SubscriberQueueName    string
}

func Load() (*Config, error) {
	isProductionStr := getEnvOrDefault("IS_PRODUCTION", defaultIsProduction)
	isProduction, err := strconv.ParseBool(isProductionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse `IS_PRODUCTION`: %w", err)
	}

	return &Config{
		IsProduction:           isProduction,
		PostgresUrl:            getEnvOrDefault("POSTGRES_URL", defaultRabbitMqUrl),
		RabbitMqUrl:            getEnvOrDefault("RABBITMQ_URL", defaultRabbitMqUrl),
		PublisherRoutingKey:    getEnvOrDefault("PUBLISHER_ROUTING_KEY", defaultPublisherRoutingKey),
		PublisherExchangeName:  getEnvOrDefault("PUBLISHER_EXCHANGE_NAME", defaultPublisherExchangeName),
		SubscriberQueueName:    getEnvOrDefault("SUBSCRIBER_QUEUE_NAME", defaultSubscriberQueueName),
		SubscriberBindingKey:   getEnvOrDefault("SUBSCRIBER_BINDING_KEY", defaultSubscriberBindingKey),
		SubscriberExchangeName: getEnvOrDefault("SUBSCRIBER_EXCHANGE_NAME", defaultSubscriberExchangeName),
	}, nil
}

func getEnvOrDefault(envVarName, defaultValue string) string {
	fromEnv, ok := os.LookupEnv(envVarName)
	if ok {
		return fromEnv
	}

	return defaultValue
}
