package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/models"
)

type OrderService interface {
	Create(ctx context.Context, req dtos.CreateOrderRequest) (*models.Order, error)

	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)

	UpdateStatusByEvent(ctx context.Context, orderID uuid.UUID, newStatus string) error
}