package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultRabbitMqUrl            = "amqp://guest:guest@localhost:5672"
	defaultIsProduction           = "false"
	defaultSubscriberExchangeName = "schedule.topic"
	defaultSubscriberBindingKey   = "schedule.parsed"
	defaultSubscriberQueueName    = "queue.service"
	defaultPublisherExchangeName  = "queue.topic"
	defaultPublisherRoutingKey    = "queue.updates.full"
)

type Config struct {
	IsProduction           bool
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
		RabbitMqUrl:            getEnvOrDefault("RABBITMQ_URL", defaultRabbitMqUrl),
		PublisherRoutingKey:    getEnvOrDefault("PUBLISHER_ROUTING_KEY", defaultPublisherRoutingKey),
		PublisherExchangeName:  getEnvOrDefault("PUBLISHER_EXCHANGE_NAME", defaultPublisherExchangeName),
		SubscriberQueueName:    getEnvOrDefault("SUBSCRIBER_QUEUE_NAME", defaultPublisherExchangeName),
		SubscriberBindingKey:   getEnvOrDefault("SUBSCRIBER_BINDING_KEY", defaultPublisherRoutingKey),
		SubscriberExchangeName: getEnvOrDefault("SUBSCRIBER_EXCHANGE_NAME", defaultPublisherExchangeName),
	}, nil
}

func getEnvOrDefault(envVarName, defaultValue string) string {
	fromEnv, ok := os.LookupEnv(envVarName)
	if ok {
		return fromEnv
	}

	return defaultValue
}
