package handler

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/broker"
	"log/slog"
)

type Handler struct {
	broker *broker.Broker
}

func New(broker *broker.Broker) *Handler {
	return &Handler{
		broker: broker,
	}
}
func (h *Handler) Handle(ctx context.Context, body string) error {
	slog.InfoContext(ctx, "received a message", "content", body)
	return nil
}
