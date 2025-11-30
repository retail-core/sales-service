package mq

import (
	"context"

	"github.com/gofrs/uuid"
)

// MessageQueuePublisher defines the contract for sending asynchronous events.
type MessageQueuePublisher interface {
	// PublishConfirmation sends an event to confirm the reservation and deduct stock permanently.
	PublishConfirmation(ctx context.Context, reservationID uuid.UUID, deductions map[string]int) error
	
	// PublishRollback sends an event to roll back a reservation (stock back to available).
	PublishRollback(ctx context.Context, orderID uuid.UUID, reservationID uuid.UUID, deductions map[string]int) error
	
	// Close cleans up the connection.
	Close()
}