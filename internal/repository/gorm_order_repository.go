package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	_ "time/tzdata"

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

func (r *GormOrderRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, from time.Time, to time.Time,
) ([]models.Order, error) {
	var orders []models.Order

	result := r.DB.WithContext(ctx).
		Where(
			"store_id = ? AND created_at >= ? AND created_at < ?",
			storeID,
			from,
			to,
		).
		Order("created_at DESC").
		Find(&orders)

	if result.Error != nil {
		return nil, fmt.Errorf(
			"failed to find orders for store ID %s: %w",
			storeID,
			result.Error,
		)
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

func (r *GormOrderRepository) GetTodayOrdersCount(
	ctx context.Context,
	storeID uuid.UUID,
) (int64, error) {

	var count int64

	now := time.Now()

	startOfDay := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		time.UTC,
	)

	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.DB.WithContext(ctx).
		Model(&models.Order{}).
		Where(
			"store_id = ? AND created_at >= ? AND created_at < ?",
			storeID,
			startOfDay,
			endOfDay,
		).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf(
			"failed to count today's orders for store %s: %w",
			storeID,
			err,
		)
	}

	return count, nil
}

func (r *GormOrderRepository) GetTodayOrders(
	ctx context.Context,
	storeID uuid.UUID,
) ([]models.Order, error) {
	var orders []models.Order

	loc, err := time.LoadLocation("Africa/Lagos")
	if err != nil {
		return nil, fmt.Errorf("failed to load store timezone: %w", err)
	}

	now := time.Now().In(loc)

	startOfDay := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0,
		loc,
	)

	startOfTomorrow := startOfDay.AddDate(0, 0, 1)

	err = r.DB.WithContext(ctx).
		Where(`
			store_id = ?
			AND status = ?
			AND created_at >= ?
			AND created_at < ?
		`,
			storeID,
			models.OrderCompleted,
			startOfDay,
			startOfTomorrow,
		).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, fmt.Errorf(
			"failed to find today's orders for store %s: %w",
			storeID,
			err,
		)
	}

	return orders, nil
}

func (r *GormOrderRepository) GetOrderSummariesByStoreID(
	ctx context.Context,
	storeID uuid.UUID,
) ([]dtos.OrderSummary, error) {

	loc, _err := time.LoadLocation("Africa/Lagos")
	if _err != nil {
		return nil, fmt.Errorf("failed to load store timezone: %w", _err)
	}

	type summaryRow struct {
		Month           int64   `gorm:"column:month"`
		Year            int64   `gorm:"column:year"`
		TotalOrderCount int64   `gorm:"column:total_order_count"`
		TotalRevenue    float64 `gorm:"column:total_revenue"`
		TotalProfit     float64 `gorm:"column:total_profit"`
	}

	var monthlyRows []summaryRow

	err := r.DB.WithContext(ctx).
		Model(&models.Order{}).
		Select(`
			EXTRACT(MONTH FROM created_at AT TIME ZONE 'Africa/Lagos')::int AS month,
			EXTRACT(YEAR FROM created_at AT TIME ZONE 'Africa/Lagos')::int AS year,
			COUNT(id) AS total_order_count,
			COALESCE(SUM(total_amount), 0) AS total_revenue,
			COALESCE(SUM(total_amount - total_cost), 0) AS total_profit
		`).
		Where(`
			store_id = ?
			AND status = ?
		`, storeID, models.OrderCompleted).
		Group(`
			EXTRACT(YEAR FROM created_at AT TIME ZONE 'Africa/Lagos'),
			EXTRACT(MONTH FROM created_at AT TIME ZONE 'Africa/Lagos')
		`).
		Order("year DESC, month DESC").
		Scan(&monthlyRows).Error

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get monthly order summaries for store %s: %w",
			storeID,
			err,
		)
	}

	summaries := make([]dtos.OrderSummary, 0, len(monthlyRows)+1)

	for _, row := range monthlyRows {

		from := time.Date(
			int(row.Year),
			time.Month(row.Month),
			1,
			0, 0, 0, 0,
			loc,
		)

		to := from.AddDate(0, 1, 0)

		summaries = append(summaries, dtos.OrderSummary{
			Type:            dtos.OrderSummaryMonth,
			Month:           int(row.Month),
			Year:            int(row.Year),
			From:            from,
			To:              to,
			TotalOrderCount: row.TotalOrderCount,
			TotalRevenue:    row.TotalRevenue,
			TotalProfit:     row.TotalProfit,
		})
	}

	now := time.Now().In(loc)

	today := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0,
		loc,
	)

	startOfYesterday := today.AddDate(0, 0, -1)

	type yesterdayRow struct {
		TotalOrderCount int64   `gorm:"column:total_order_count"`
		TotalRevenue    float64 `gorm:"column:total_revenue"`
		TotalProfit     float64 `gorm:"column:total_profit"`
	}

	var yesterday yesterdayRow

	err = r.DB.WithContext(ctx).
		Model(&models.Order{}).
		Select(`
			COUNT(id) AS total_order_count,
			COALESCE(SUM(total_amount), 0) AS total_revenue,
			COALESCE(SUM(total_amount - total_cost), 0) AS total_profit
		`).
		Where(`
			store_id = ?
			AND status = ?
			AND created_at >= ?
			AND created_at < ?
		`,
			storeID,
			models.OrderCompleted,
			startOfYesterday,
			today,
		).
		Scan(&yesterday).Error

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get yesterday's order summary for store %s: %w",
			storeID,
			err,
		)
	}

	summaries = append(
		[]dtos.OrderSummary{
			{
				Type:            dtos.OrderSummaryYesterday,
				Month:           int(startOfYesterday.Month()),
				Year:            startOfYesterday.Year(),
				From:            startOfYesterday,
				To:              today,
				TotalOrderCount: yesterday.TotalOrderCount,
				TotalRevenue:    yesterday.TotalRevenue,
				TotalProfit:     yesterday.TotalProfit,
			},
		},
		summaries...,
	)

	return summaries, nil
}
