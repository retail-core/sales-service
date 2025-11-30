package mq

import (
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func InitRabbitMQ(amqpURI string) (*RabbitMQConnection, error) {

	config := amqp091.Config{
		Properties: amqp091.Table{
			"connection_name": "sales-service",
		},
	}

	conn, err := amqp091.DialConfig(amqpURI, config)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQConnection{conn: conn, channel: channel}, nil
}