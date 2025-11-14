package clients

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Close()
	Publish(routing string, payload any) error
}

type Rabbit struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	exch    string
}

func MustRabbit(url, exchange, exchangeType string) *Rabbit {
	conn, err := amqp091.Dial(url)
	if err != nil {
		log.Fatalf("rabbit dial: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("rabbit channel: %v", err)
	}
	if err := ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		log.Fatalf("exchange declare: %v", err)
	}
	return &Rabbit{conn: conn, channel: ch, exch: exchange}
}

func (r *Rabbit) Close() { r.channel.Close(); r.conn.Close() }

func (r *Rabbit) GetChannel() *amqp091.Channel {
	return r.channel
}

func (r *Rabbit) Publish(routing string, payload any) error {
	b, _ := json.Marshal(payload)
	return r.channel.Publish(r.exch, routing, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        b,
	})
}
