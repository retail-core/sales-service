package repository

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID) ([]models.Order, error)

	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	GetSalesReport(ctx context.Context, storeID uuid.UUID, from, to time.Time) (*dtos.SalesReportResponse, error)
	GetTodayOrdersCount(ctx context.Context, storeID uuid.UUID) (int64, error)
}