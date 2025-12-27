package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	"github.com/retail-core/sales-service/internal/dtos"
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

	// no need to preload items here, can be added if necessary, // also sort by created_at desc so that latest orders come first
	result := r.DB.WithContext(ctx).Where("store_id = ?", storeID).Order("created_at desc").Find(&orders)
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

func (r *GormOrderRepository) GetSalesReport(
	ctx context.Context,
	storeID uuid.UUID,
	from, to time.Time,
) (*dtos.SalesReportResponse, error) {

	report := &dtos.SalesReportResponse{
		StoreID:        storeID,
		From:           from,
		To:             to,
		PaymentMethods: make(map[string]float64),
	}

	// 1️⃣ Total sales + total orders
	type summaryRow struct {
		TotalSales  float64
		TotalOrders int
	}

	var summary summaryRow

	err := r.DB.WithContext(ctx).
		Model(&models.Order{}).
		Select(`
			COALESCE(SUM(total_amount), 0) AS total_sales,
			COUNT(id) AS total_orders
		`).
		Where(`
			store_id = ?
			AND status = ?
			AND created_at >= ?
			AND created_at < ?
		`, storeID, models.OrderCompleted, from, to).
		Scan(&summary).Error

	if err != nil {
		return nil, err
	}

	report.TotalSales = summary.TotalSales
	report.TotalOrders = summary.TotalOrders

	// 2️⃣ Best selling items
	err = r.DB.WithContext(ctx).
		Table("order_items").
		Select(`
			order_items.inventory_id,
			order_items.product_name,
			SUM(order_items.quantity) AS quantity_sold,
			SUM(order_items.subtotal) AS total_revenue,
			SUM(
				(order_items.unit_price - COALESCE(order_items.cost_price, 0))
				* order_items.quantity
			) AS profit
		`).
		Joins(`
			JOIN orders ON orders.id = order_items.order_id
		`).
		Where(`
			orders.store_id = ?
			AND orders.status = ?
			AND orders.created_at >= ?
			AND orders.created_at < ?
		`, storeID, models.OrderCompleted, from, to).
		Group(`
			order_items.inventory_id,
			order_items.product_name
		`).
		Order("quantity_sold DESC").
		Limit(5).
		Scan(&report.BestSellingItems).Error

	if err != nil {
		return nil, err
	}

	// 3️⃣ Profit (overall)
	type profitRow struct {
		Profit float64
	}

	var profit profitRow

	err = r.DB.WithContext(ctx).
		Table("order_items").
		Select(`
			COALESCE(SUM(
				(order_items.unit_price - COALESCE(order_items.cost_price, 0))
				* order_items.quantity
			), 0) AS profit
		`).
		Joins(`
			JOIN orders ON orders.id = order_items.order_id
		`).
		Where(`
			orders.store_id = ?
			AND orders.status = ?
			AND orders.created_at >= ?
			AND orders.created_at < ?
		`, storeID, models.OrderCompleted, from, to).
		Scan(&profit).Error

	if err != nil {
		return nil, err
	}

	report.Profit = profit.Profit

	// 4️⃣ Payment method breakdown
	type paymentRow struct {
		Method string
		Amount float64
	}

	var payments []paymentRow

	err = r.DB.WithContext(ctx).
		Model(&models.Order{}).
		Select(`
			payment_method AS method,
			SUM(total_amount) AS amount
		`).
		Where(`
			store_id = ?
			AND status = ?
			AND created_at >= ?
			AND created_at < ?
		`, storeID, models.OrderCompleted, from, to).
		Group("payment_method").
		Scan(&payments).Error

	if err != nil {
		return nil, err
	}

	for _, p := range payments {
		report.PaymentMethods[p.Method] = p.Amount
	}

	return report, nil
}
