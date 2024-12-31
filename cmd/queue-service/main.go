package main

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/broker"
	"electricity-schedule-bot/queue-service/internal/logger"
	"log/slog"
	"os"
)

func main() {
	var handler slog.Handler
	handler = logger.NewTraceIdHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	logger := slog.New(handler)
	logger = logger.With("service", "queue-service")
	slog.SetDefault(logger)

	broker, err := broker.New(broker.BrokerConfig{
		RabbitMQUrl: "amqp://guest:guest@localhost:5672",
		Publisher: broker.PublisherConfig{
			ExchangeName: "updates.topic",
			RoutingKey:   "updates.full",
		},
		Subscriber: broker.SubscriberConfig{
			ExchangeName: "schedule.topic",
			BindingKey:   "schedule.parsed",
			QueueName:    "queue.service",
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
