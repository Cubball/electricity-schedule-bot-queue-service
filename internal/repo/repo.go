package repo

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/models"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type RepoConfig struct {
	PostgresUrl string
}

type Repo struct {
	postgresUrl string
	db          *pgx.Conn
}

func New(config RepoConfig) (*Repo, error) {
	db, err := pgx.Connect(context.Background(), config.PostgresUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to pg: %w", err)
	}

	return &Repo{
		postgresUrl: config.PostgresUrl,
		db:          db,
	}, nil
}

func (r *Repo) GetAllQueues() ([]models.Queue, error) {
	return nil, nil
}

func (r *Repo) UpdateAllQueues(queues []models.Queue) error {
    return nil
}
