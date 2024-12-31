package handler

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/broker"
	"electricity-schedule-bot/queue-service/internal/models"
	"electricity-schedule-bot/queue-service/internal/repo"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"time"
)

// HACK: ignoring the possible error, it shouldn't
var location, _ = time.LoadLocation("Europe/Kyiv")

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

func (h *Handler) Handle(ctx context.Context, body []byte) error {
	slog.InfoContext(ctx, "received a message")
	var schedule models.Schedule
	err := json.Unmarshal(body, &schedule)
	if err != nil {
		return fmt.Errorf("failed to unmarshal the message: %w", err)
	}

	slog.DebugContext(ctx, "got queues from the schedule", "queuesCount", len(schedule.Queues), "fetchTime", schedule.FetchTime)
	queues, err := h.repo.GetAllQueues(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch the queues from the db: %w", err)
	}

	slog.DebugContext(ctx, "fetched queues from the db", "queuesCount", len(queues))
	updatedQueues, err := getUpdatedQueues(queues, schedule.Queues, schedule.FetchTime)
	if err != nil {
		return fmt.Errorf("failed to determine the updated queues: %w", err)
	}

	slog.DebugContext(ctx, "got updated queues", "queuesCount", len(queues))
	err = h.broker.Publish(ctx, models.ScheduleUpdate{
		FetchTime:     schedule.FetchTime,
		UpdatedQueues: updatedQueues,
	})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	err = h.repo.UpdateAllQueues(updatedQueues)
	if err != nil {
		return fmt.Errorf("failed to update the queues: %w", err)
	}

    slog.DebugContext(ctx, "updated queues in the db")
	return nil
}

func getUpdatedQueues(old, new []models.Queue, fetchTime time.Time) ([]models.Queue, error) {
	updatedQueues := []models.Queue{}
	for _, newQueue := range new {
		oldQueueIdx := slices.IndexFunc(old, func(oldQueue models.Queue) bool {
			return oldQueue.Number == newQueue.Number
		})
		if oldQueueIdx == -1 {
			return nil, fmt.Errorf("failed to find an old queue with number %q", newQueue.Number)
		}

		oldQueue := old[oldQueueIdx]
		removed, added := diff(oldQueue.DisconnectionTimes, newQueue.DisconnectionTimes)
		if len(added) > 0 {
			updatedQueues = append(updatedQueues, newQueue)
			continue
		}

		for _, removedTime := range removed {
			if removedTime.End.Compare(fetchTime) >= 0 {
				updatedQueues = append(updatedQueues, newQueue)
				break
			}
		}
	}

	return updatedQueues, nil
}

func diff(old, new []models.DisconnectionTime) (removed, added []models.DisconnectionTime) {
	for _, oldTime := range old {
		if !slices.ContainsFunc(new, func(newTime models.DisconnectionTime) bool {
			return oldTime.Start.Equal(newTime.Start) && oldTime.End.Equal(newTime.End)
		}) {
			removed = append(removed, oldTime)
		}
	}

	for _, newTime := range new {
		if !slices.ContainsFunc(old, func(oldTime models.DisconnectionTime) bool {
			return oldTime.Start.Equal(newTime.Start) && oldTime.End.Equal(newTime.End)
		}) {
			added = append(added, newTime)
		}
	}

	return
}
