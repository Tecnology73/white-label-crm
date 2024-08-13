package rabbitmq

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"reflect"
)

var (
	client  *amqp.Connection
	channel *amqp.Channel
)

func NewConnection(url string, opts *amqp.Config, prefetchCount int) *amqp.Connection {
	var err error
	client, err = amqp.DialConfig(url, *opts)
	if err != nil {
		log.Fatalf("[rabbitmq.NewConnection] %v\n", err)
	}

	channel, err = client.Channel()
	if err != nil {
		log.Fatalf("[rabbitmq.NewConnection] %v\n", err)
	}

	if err = channel.Qos(prefetchCount, 0, true); err != nil {
		log.Fatalf("[rabbitmq.NewConnection] %v\n", err)
	}

	return client
}

func CloseConnection() {
	if err := channel.Close(); err != nil {
		log.Fatalf("[rabbitmq.CloseConnection] %v\n", err)
	}
	channel = nil

	if err := client.Close(); err != nil {
		log.Fatalf("[rabbitmq.CloseConnection] %v\n", err)
	}
	client = nil
}

type AckFunc = func(multiple bool) error
type NackFunc = func(multiple bool, requeue bool) error

type Event interface {
	EventName() string
}

type BrandUpdatedEvent struct {
}

func (e *BrandUpdatedEvent) EventName() string { return "BrandUpdated" }

func Listen[T Event](cb func(event T, ack AckFunc, nack NackFunc)) error {
	msgs, err := channel.Consume(
		"",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			var event T
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("[rabbitmq.Listen[%v]] Unmarshal | %v\n", reflect.TypeFor[T](), err)
				if err = msg.Nack(false, true); err != nil {
					log.Printf("[rabbitmq.Listen[%v]] Nack | %v\n", reflect.TypeFor[T](), err)
				}
			}

			cb(event, msg.Ack, msg.Nack)
		}
	}()

	return nil
}

func Publish[T Event](event T) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return PublishInternal(event.EventName(), body)
}

func PublishInternal(queue string, body []byte) error {
	q, err := channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return channel.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			Headers:      amqp.Table{},
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}
