package service

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/models"
)

type OrderService interface {
	Create(ctx context.Context, storeID uuid.UUID, req dtos.CreateOrderRequest) (*models.Order, error)

	GetOrderByID(ctx context.Context, storeID uuid.UUID, orderID uuid.UUID) (*models.Order, error)
	GetOrdersByStoreID(ctx context.Context, storeID uuid.UUID) ([]models.Order, error)

	UpdateStatusByEvent(ctx context.Context, orderID uuid.UUID, newStatus string) error
}