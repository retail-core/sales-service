package mq

import (
	"context"

	"github.com/gofrs/uuid"
)

// MessageQueuePublisher defines the contract for sending asynchronous events.
type MessageQueuePublisher interface {
	PublishConfirmation(ctx context.Context, reservationID uuid.UUID) error
	
	PublishRollback(ctx context.Context, reservationID uuid.UUID) error
	
	Close()
}