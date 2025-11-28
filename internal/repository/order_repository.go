package repository

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	
	FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error)

	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}