package broker

import (
	"context"
	"electricity-schedule-bot/queue-service/internal/logger"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	traceIdHeaderName      = "X-Trace-Id"
	reconnectWaitInSeconds = 10
	contentType            = "application/json"
)

type BrokerConfig struct {
	RabbitMQUrl string
	Publisher   PublisherConfig
	Subscriber  SubscriberConfig
}

type PublisherConfig struct {
	ExchangeName string
	RoutingKey   string
}

type SubscriberConfig struct {
	ExchangeName string
	BindingKey   string
	QueueName    string
}

type Broker struct {
	config           BrokerConfig
	connection       *amqp.Connection
	channel          *amqp.Channel
	closeChannel     chan *amqp.Error
	reconnectChannel chan bool
}

func New(config BrokerConfig) (*Broker, error) {
	broker := Broker{
		config:           config,
		reconnectChannel: make(chan bool, 1),
	}
	err := broker.connect()
	if err != nil {
		return nil, err
	}

	broker.handleReconnect()
	return &broker, nil
}

func (p *Broker) Publish(ctx context.Context, obj any) error {
	body, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal an object to json: %w", err)
	}

	traceId, ok := ctx.Value(logger.TraceIdContextKey).(string)
	if !ok {
		traceId = uuid.NewString()
		slog.WarnContext(ctx, "failed to get the trace id from the context, generating a new one", "traceId", traceId)
	}

	headers := amqp.Table{
		traceIdHeaderName: traceId,
	}
	slog.DebugContext(ctx, "will publish a message to rmq", "content", string(body))
	err = p.channel.Publish(p.config.Publisher.ExchangeName, p.config.Publisher.RoutingKey, false, false, amqp.Publishing{
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	slog.DebugContext(ctx, "published a message to rmq")
	return nil
}

func (p *Broker) RegisterSubscriber(subscriber func(context.Context, string) error) error {
	// doing autoAck for now for simplicity
	msgChannel, err := p.channel.Consume(p.config.Subscriber.QueueName, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	slog.Debug("started consuming messages")
	go func() {
		for {
			for msg := range msgChannel {
				handleMessage(msg, subscriber)
			}

			<-p.reconnectChannel
			msgChannel, err = p.channel.Consume(p.config.Subscriber.QueueName, "", true, false, false, false, nil)
			if err != nil {
				// TODO: handle this better
				slog.Error("failed to reregister subscriber", "err", err)
				break
			}
		}
	}()

	return nil
}

func (p *Broker) Close() {
	p.channel.Close()
	p.connection.Close()
}

func (p *Broker) connect() error {
	connection, err := amqp.Dial(p.config.RabbitMQUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to rmq: %w", err)
	}

	p.connection = connection
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	p.channel = channel
	err = channel.ExchangeDeclare(p.config.Publisher.ExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange for publisher: %w", err)
	}

	err = channel.ExchangeDeclare(p.config.Subscriber.ExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange for subscriber: %w", err)
	}

	_, err = channel.QueueDeclare(p.config.Subscriber.QueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = channel.QueueBind(p.config.Subscriber.QueueName, p.config.Subscriber.BindingKey, p.config.Subscriber.ExchangeName, false, nil)
	if err != nil {
		return fmt.Errorf("failed to bind a queue to an exchange: %w", err)
	}

	p.closeChannel = connection.NotifyClose(make(chan *amqp.Error))
	slog.Info("connected to rmq")
	return nil
}

func (p *Broker) handleReconnect() {
	go func() {
		for {
			err := <-p.closeChannel
			if err == nil {
				break
			}

			slog.Warn("disconnected from rmq")
			for {
				slog.Info("trying to reconnect to rmq")
				err := p.connect()
				if err == nil {
					p.reconnectChannel <- true
					break
				}

				time.Sleep(time.Second * reconnectWaitInSeconds)
			}
		}
	}()
}

func handleMessage(msg amqp.Delivery, subscriber func(context.Context, string) error) {
	body := string(msg.Body)
	traceId, ok := msg.Headers[traceIdHeaderName].(string)
	ctx := context.Background()
	if !ok {
		traceId = uuid.NewString()
		slog.WarnContext(ctx, "failed to get the trace id from the message headers, generating a new one", "traceId", traceId)
	}

	ctx = context.WithValue(ctx, logger.TraceIdContextKey, traceId)
	err := subscriber(ctx, body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to process a message", "err", err)
	}
}
