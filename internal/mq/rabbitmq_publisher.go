package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/retail-core/sales-service/internal/logger"
	"go.uber.org/zap"
)

const (
	// Exchange name for inventory events (Adjust as needed)
	InventoryExchange = "inventory_events"
	// Routing key for stock confirmation
	ConfirmationKey = "inventory.reservation.confirm"
	// Routing key for stock rollback
	RollbackKey = "inventory.reservation.rollback"
)

// NewRabbitMQPublisher establishes the connection and channel.
func DeclareExchange(mq *RabbitMQConnection, exchangeName string) error {

	err := mq.channel.ExchangeDeclare(
		InventoryExchange, // name
		"topic",           // type (topic is flexible for routing keys)
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		mq.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	logger.L().Info("✅ Connected to RabbitMQ and exchange declared", zap.String("exchange", exchangeName))
	return nil
}

// publish sends the message to the specified routing key.
func (p *RabbitMQConnection) publish(ctx context.Context, routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Use the context for a timeout when publishing
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(ctx,
		InventoryExchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent, // Message survives a broker restart
			Timestamp:    time.Now(),
		})
}

func (p *RabbitMQConnection) PublishConfirmation(ctx context.Context, reservationID uuid.UUID) error {
	event := struct {
		ReservationID uuid.UUID `json:"reservation_id"`
	}{
		ReservationID: reservationID,
	}
	logger.L().Info("Publishing confirmation event", zap.String("reservation_id", reservationID.String()))
	return p.publish(ctx, ConfirmationKey, event)
}

func (p *RabbitMQConnection) PublishRollback(ctx context.Context, reservationID uuid.UUID) error {
	event := struct {
		ReservationID uuid.UUID `json:"reservation_id"`
	}{
		ReservationID: reservationID,
	}
	logger.L().Info("Publishing rollback event", zap.String("reservation_id", reservationID.String()))
	return p.publish(ctx, RollbackKey, event)
}

func (p *RabbitMQConnection) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	log.Println("RabbitMQ connection closed.")
}
