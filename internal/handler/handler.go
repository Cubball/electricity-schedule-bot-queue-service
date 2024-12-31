package handler

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/broker"
	"electricity-schedule-bot/queue-service/internal/repo"
	"log/slog"
)

type Handler struct {
	broker *broker.Broker
	repo   *repo.Repo
}

func New(broker *broker.Broker, repo *repo.Repo) *Handler {
	return &Handler{
		broker: broker,
		repo:   repo,
	}
}
func (h *Handler) Handle(ctx context.Context, body string) error {
	slog.InfoContext(ctx, "received a message", "content", body)
	return nil
}
