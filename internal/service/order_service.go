package service

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/models"
	"github.com/retail-core/sales-service/internal/redis_client"
)

type OrderService interface {
	Create(ctx context.Context, storeID uuid.UUID, req dtos.CreateOrderRequest) (*models.Order, int64, error)

	GetOrderByID(ctx context.Context, storeID uuid.UUID, orderID uuid.UUID) (*models.Order, error)

	UpdateStatusByEvent(ctx context.Context, orderID uuid.UUID, newStatus string) error
	GetSalesReport(ctx context.Context, storeID uuid.UUID, from, to time.Time) (*dtos.SalesReportResponse, error)
	GetQueueStoreClient() redis_client.QueueStore

	GetOrdersByStoreID(ctx context.Context, storeID uuid.UUID, from, to time.Time) ([]models.Order, error)
	GetDashboardByStoreID(ctx context.Context, storeID uuid.UUID) ([]models.Order, []dtos.OrderSummary, error)
}