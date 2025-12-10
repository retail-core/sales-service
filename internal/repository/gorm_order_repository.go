package repository

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	"github.com/retail-core/sales-service/internal/models"
)

type GormOrderRepository struct {
	DB *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{DB: db}
}

func (r *GormOrderRepository) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	if result := r.DB.WithContext(ctx).Create(order); result.Error != nil {
		return nil, fmt.Errorf("failed to create order and items: %w", result.Error)
	}
	return order, nil
}

func (r *GormOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	order := &models.Order{}
	
	result := r.DB.WithContext(ctx).Preload("Items").First(order, id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Or a custom models.ErrNotFound error
		}
		return nil, fmt.Errorf("failed to find order by ID %s: %w", id, result.Error)
	}
	return order, nil
}

func (r *GormOrderRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID) ([]models.Order, error) {
	var orders []models.Order

	// no need to preload items here, can be added if necessary
	result := r.DB.WithContext(ctx).Where("store_id = ?", storeID).Find(&orders)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find orders for store ID %s: %w", storeID, result.Error)
	}
	return orders, nil
}

func (r *GormOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result := r.DB.WithContext(ctx).Model(&models.Order{Base: models.Base{ID: id}}).UpdateColumns(map[string]interface{}{
		"status": status,
	})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update order status for ID %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("order not found or status already set for ID %s", id)
	}
	
	return nil
}
