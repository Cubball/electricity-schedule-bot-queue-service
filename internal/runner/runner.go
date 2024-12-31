package runner

import (
	"electricity-schedule-bot/queue-service/internal/broker"
	"electricity-schedule-bot/queue-service/internal/handler"
	"fmt"
	"log/slog"
)

type Runner struct {
	handler     *handler.Handler
	broker      *broker.Broker
	stopChannel chan bool
}

func New(handler *handler.Handler, broker *broker.Broker) *Runner {
	return &Runner{
		handler:     handler,
		broker:      broker,
		stopChannel: make(chan bool),
	}
}

func (r *Runner) Run() error {
	err := r.broker.RegisterSubscriber(r.handler.Handle)
	if err != nil {
		return fmt.Errorf("failed to start the runner %w", err)
	}

	slog.Info("the runner has started")
	return nil
}

func (r *Runner) Wait() {
	<-r.stopChannel
}

func (r *Runner) Stop() {
	r.stopChannel <- true
	close(r.stopChannel)
	r.broker.Close()
}
