package main

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/broker"
	"electricity-schedule-bot/queue-service/internal/config"
	"electricity-schedule-bot/queue-service/internal/logger"
	"log/slog"
	"os"
)

func main() {
	config, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		return
	}

	var handler slog.Handler
	if config.IsProduction {
		handler = logger.NewTraceIdHandler(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		handler = logger.NewTraceIdHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	logger := slog.New(handler)
	logger = logger.With("service", "queue-service")
	slog.SetDefault(logger)

	broker, err := broker.New(broker.BrokerConfig{
		RabbitMQUrl: config.RabbitMqUrl,
		Publisher: broker.PublisherConfig{
			ExchangeName: config.PublisherExchangeName,
			RoutingKey:   config.PublisherRoutingKey,
		},
		Subscriber: broker.SubscriberConfig{
			ExchangeName: config.SubscriberExchangeName,
			BindingKey:   config.SubscriberBindingKey,
			QueueName:    config.SubscriberQueueName,
		},
	})
	if err != nil {
		slog.Error("failed to init broker", "err", err)
	}

	broker.RegisterSubscriber(func(ctx context.Context, body string) error {
		slog.InfoContext(ctx, "received a message", "content", body)
		return nil
	})

	forever := make(chan bool)
	<-forever
}
