package repo

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/models"
	"fmt"
	"strings"
	"time"

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

func (r *Repo) UpdateAllQueues(ctx context.Context, queues []models.Queue) error {
	transaction, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin a transaction: %w", err)
	}

	defer transaction.Rollback(ctx)
	deleteQuery := "DELETE FROM disconnection_times;"
	_, err = transaction.Exec(ctx, deleteQuery)
	if err != nil {
		return fmt.Errorf("failed to execute DELETE within a transaction: %w", err)
	}

	for _, queue := range queues {
		for _, time := range queue.DisconnectionTimes {
			_, err = transaction.Exec(ctx, `
                INSERT INTO disconnection_times (queue_number, start_time, end_time)
                VALUES ($1, $2, $3);
            `, queue.Number, time.Start, time.End)
			if err != nil {
				return fmt.Errorf("failed to execute INSERT within a transaction: %w", err)
			}
		}
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit a transaction: %w", err)
	}

	return nil
}

func (r *Repo) GetAllQueues(ctx context.Context) ([]models.Queue, error) {
	query := `
        SELECT q.number, dt.start_time, dt.end_time
        FROM queues AS q
        LEFT JOIN disconnection_times AS dt
        ON q.number = dt.queue_number;
    `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the queues from db: %w", err)
	}

	defer rows.Close()
	queueMap := map[string]*models.Queue{}
	for rows.Next() {
		var queueNumber string
		var start *time.Time
		var end *time.Time
		err = rows.Scan(&queueNumber, &start, &end)
		if err != nil {
			return nil, fmt.Errorf("failed to read the record from db: %w", err)
		}

		queue, ok := queueMap[queueNumber]
		if !ok {
			queue = &models.Queue{
				Number:             strings.TrimSpace(queueNumber),
				DisconnectionTimes: []models.DisconnectionTime{},
			}
			queueMap[queueNumber] = queue
		}

		if start != nil && end != nil {
			queue.DisconnectionTimes = append(queue.DisconnectionTimes, models.DisconnectionTime{
				Start: *start,
				End:   *end,
			})
		}
	}

	queues := []models.Queue{}
	for _, v := range queueMap {
		queues = append(queues, *v)
	}

	return queues, nil
}

func (r *Repo) Close() {
	r.db.Close(context.Background())
}
