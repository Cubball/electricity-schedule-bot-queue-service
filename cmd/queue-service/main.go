package main

import (
	"electricity-schedule-bot/queue-service/internal/broker"
	"electricity-schedule-bot/queue-service/internal/config"
	"electricity-schedule-bot/queue-service/internal/handler"
	"electricity-schedule-bot/queue-service/internal/logger"
	"electricity-schedule-bot/queue-service/internal/repo"
	"electricity-schedule-bot/queue-service/internal/runner"
	"log/slog"
	"os"
)

func main() {
	config, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		return
	}

	var slogHandler slog.Handler
	if config.IsProduction {
		slogHandler = logger.NewTraceIdHandler(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		slogHandler = logger.NewTraceIdHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	logger := slog.New(slogHandler)
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

    repo, err := repo.New(repo.RepoConfig{
        PostgresUrl: config.PostgresUrl,
    })
	if err != nil {
		slog.Error("failed to init repo", "err", err)
	}

	handler := handler.New(broker, repo)
	runner := runner.New(handler, broker)
	err = runner.Run()
	if err != nil {
		slog.Error("failed to run the runner", "err", err)
		return
	}

	runner.Wait()
}
