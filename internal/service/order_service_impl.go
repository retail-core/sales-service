package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/client"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/errors"
	"github.com/retail-core/sales-service/internal/logger"
	"github.com/retail-core/sales-service/internal/models"
	"github.com/retail-core/sales-service/internal/mq"
	"github.com/retail-core/sales-service/internal/redis_client"
	"github.com/retail-core/sales-service/internal/repository"
	"go.uber.org/zap"
	// TODO: Inventory service client import will go here (e.g., github.com/your-username/inventory-client)
	// TODO: RabbitMQ client import will go here (e.g., github.com/streadway/amqp)
)

// Implementation Note: In a real microservice, we would inject interfaces for
// the InventoryClient and MessageQueuePublisher here as well.

// OrderServiceImpl is the concrete implementation of OrderService.
type OrderServiceImpl struct {
	OrderRepo       repository.OrderRepository
	InventoryClient client.InventoryClient
	MQPublisher     mq.MessageQueuePublisher
	RedisQueueStore redis_client.QueueStore
}

func NewOrderServiceImpl(repo repository.OrderRepository, inventoryClient client.InventoryClient, mqPublisher mq.MessageQueuePublisher, redisQueueStore redis_client.QueueStore) *OrderServiceImpl {
	return &OrderServiceImpl{
		OrderRepo:       repo,
		InventoryClient: inventoryClient,
		MQPublisher:     mqPublisher,
		RedisQueueStore: redisQueueStore,
	}
}

func (s *OrderServiceImpl) Create(
	ctx context.Context,
	storeID uuid.UUID,
	req dtos.CreateOrderRequest,
) (*models.Order, int64, error) {

	res, err := s.InventoryClient.ReserveAndGetSnapshot(ctx, storeID, req)
	if err != nil {
		return nil, 0, err
	}

	rollback := func(err error) (*models.Order, int64, error) {
		s.MQPublisher.PublishRollback(ctx, res.ReservationID)
		return nil, 0, err
	}

	// =========================
	// Build lookup map
	// =========================
	inventoryMap := make(map[string]client.InventorySnapshot)
	for _, s := range res.InventorySnapshots {
		key := s.InventoryID + "|" + s.UnitID
		inventoryMap[key] = s
	}

	orderID := uuid.Must(uuid.NewV4())

	order := &models.Order{
		Base:          models.Base{ID: orderID},
		StoreID:       storeID,
		CustomerName:  req.CustomerName,
		SoldBy:        req.SoldBy,
		PaymentMethod: models.PaymentMethod(req.PaymentMethod),
		Channel:       models.OrderChannelInStore,
		Status:        models.OrderCompleted,
		Items:         make([]models.OrderItem, 0),
	}

	var totalAmount float64
	var totalCost float64

	// =========================
	// INVENTORY ITEMS
	// =========================
	for _, reqItem := range req.InventoryItems {

		key := reqItem.InventoryID.String() + "|" + reqItem.UnitID.String()

		snap, ok := inventoryMap[key]
		if !ok {
			return rollback(fmt.Errorf("inventory snapshot not found: %s", reqItem.InventoryID))
		}

		cost := 0.0
		if snap.CostPrice != nil {
			cost = *snap.CostPrice
		}

		subtotal := float64(reqItem.Quantity) * snap.UnitPrice
		subtotalCost := float64(reqItem.Quantity) * float64(snap.QtyPerUnit) * cost

		totalAmount += subtotal
		totalCost += subtotalCost

		isBaseUnit := snap.IsBaseUnit

		order.Items = append(order.Items, models.OrderItem{
			Base:         models.Base{ID: uuid.Must(uuid.NewV4())},
			InventoryID:  reqItem.InventoryID,
			OrderID:      orderID,
			ProductName:  snap.Name,
			ImageUrl:     &snap.ImageUrl,
			UnitPrice:    snap.UnitPrice,
			CostPrice:    &cost,
			Quantity:     reqItem.Quantity,
			Subtotal:     subtotal,
			SubtotalCost: subtotalCost,

			UnitLabel:  snap.UnitLabel,
			QtyPerUnit: snap.QtyPerUnit,
			IsBaseUnit: &isBaseUnit,
		})
	}

	order.TotalAmount = totalAmount
	order.TotalCost = totalCost

	// =========================
	// SAVE
	// =========================
	saved, err := s.OrderRepo.CreateOrder(ctx, order)
	if err != nil {
		return rollback(fmt.Errorf("failed to save order: %w", err))
	}

	if err := s.MQPublisher.PublishConfirmation(ctx, res.ReservationID); err != nil {
		logger.L().Warn("failed to publish confirmation", zap.Error(err))
	}

	count, err := s.OrderRepo.GetTodayOrdersCount(ctx, storeID)
	if err != nil {
		logger.L().Warn("failed to get today's orders count", zap.Error(err))
	}

	return saved, count, nil
}

func (s *OrderServiceImpl) Create002(
	ctx context.Context,
	storeID uuid.UUID,
	req dtos.CreateOrderRequest,
) (*models.Order, int64, error) {

	res, err := s.InventoryClient.ReserveAndGetSnapshot(ctx, storeID, req)
	if err != nil {
		return nil, 0, err
	}

	rollback := func(err error) (*models.Order, int64, error) {
		s.MQPublisher.PublishRollback(ctx, res.ReservationID)
		return nil, 0, err
	}

	// =========================
	// Build lookup maps
	// =========================
	inventoryMap := make(map[string]client.InventorySnapshot)
	for _, s := range res.InventorySnapshots {
		inventoryMap[s.InventoryID] = s
	}

	// comboMap := make(map[string]client.ComboSnapshot)
	// for _, c := range [1, 3] {
	// 	comboMap[c.ComboID] = c
	// }

	orderID := uuid.Must(uuid.NewV4())

	order := &models.Order{
		Base:          models.Base{ID: orderID},
		StoreID:       storeID,
		CustomerName:  req.CustomerName,
		SoldBy:        req.SoldBy,
		PaymentMethod: models.PaymentMethod(req.PaymentMethod),
		Channel:       models.OrderChannelInStore,
		Status:        models.OrderCompleted,
		Items:         make([]models.OrderItem, 0),
	}

	var totalAmount float64
	var totalCost float64

	// =========================
	// INVENTORY ITEMS
	// =========================
	for _, reqItem := range req.InventoryItems {

		snap, ok := inventoryMap[reqItem.InventoryID.String()]
		if !ok {
			return rollback(fmt.Errorf("inventory snapshot not found: %s", reqItem.InventoryID))
		}

		cost := 0.0
		if snap.CostPrice != nil {
			cost = *snap.CostPrice
		}

		subtotal := float64(reqItem.Quantity) * snap.UnitPrice
		subtotalCost := float64(reqItem.Quantity) * cost

		totalAmount += subtotal
		totalCost += subtotalCost

		order.Items = append(order.Items, models.OrderItem{
			Base:        models.Base{ID: uuid.Must(uuid.NewV4())},
			InventoryID: reqItem.InventoryID,
			OrderID:     orderID,
			ProductName: snap.Name,
			ImageUrl:    &snap.ImageUrl,
			UnitPrice:   snap.UnitPrice,
			CostPrice:   &cost,
			Quantity:    reqItem.Quantity,
			Subtotal:    subtotal,
		})
	}

	// =========================
	// COMBO ITEMS
	// =========================
	// for _, comboReq := range [] {

	// 	snap, ok := comboMap[comboReq.ComboID.String()]
	// 	if !ok {
	// 		return rollback(fmt.Errorf("combo snapshot not found: %s", comboReq.ComboID))
	// 	}

	// 	comboID, _ := uuid.FromString(snap.ComboID)

	// 	for _, item := range snap.Items {

	// 		inventoryID, err := uuid.FromString(item.InventoryID)
	// 		if err != nil {
	// 			return rollback(errors.BadRequest("invalid inventory_id"))
	// 		}

	// 		cost := 0.0
	// 		if item.CostPrice != nil {
	// 			cost = *item.CostPrice
	// 		}

	// 		qty := comboReq.Quantity * int(item.Quantity)

	// 		subtotal := float64(qty) * item.UnitPrice
	// 		subtotalCost := float64(qty) * cost

	// 		totalAmount += subtotal
	// 		totalCost += subtotalCost

	// 		order.Items = append(order.Items, models.OrderItem{
	// 			Base:         models.Base{ID: uuid.Must(uuid.NewV4())},
	// 			InventoryID:  inventoryID,
	// 			OrderID:      orderID,
	// 			ProductName:  item.Name,
	// 			ImageUrl:     &item.ImageUrl,
	// 			UnitPrice:    item.UnitPrice,
	// 			CostPrice:    &cost,
	// 			Quantity:     qty,
	// 			Subtotal:     subtotal,
	// 			SubtotalCost: subtotalCost,
	// 			ComboID:      &comboID,
	// 			ComboName:    &snap.Name,
	// 		})
	// 	}
	// }

	order.TotalAmount = totalAmount
	order.TotalCost = totalCost

	// =========================
	// SAVE
	// =========================
	saved, err := s.OrderRepo.CreateOrder(ctx, order)
	if err != nil {
		return rollback(fmt.Errorf("failed to save order: %w", err))
	}

	if err := s.MQPublisher.PublishConfirmation(ctx, res.ReservationID); err != nil {
		logger.L().Warn("failed to publish confirmation", zap.Error(err))
	}

	count, err := s.OrderRepo.GetTodayOrdersCount(ctx, storeID)
	if err != nil {
		logger.L().Warn("failed to get today's orders count", zap.Error(err))
	}

	return saved, count, nil
}

func (s *OrderServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	return s.OrderRepo.GetByID(ctx, id)
}

func (s *OrderServiceImpl) UpdateStatusByEvent(ctx context.Context, orderID uuid.UUID, newStatus string) error {
	// Simple validation to ensure the status is a valid change.
	if newStatus == "COMPLETED" || newStatus == "STOCK_ERROR" {
		return s.OrderRepo.UpdateStatus(ctx, orderID, newStatus)
	}
	return fmt.Errorf("invalid status provided for update: %s", newStatus)
}

func (s *OrderServiceImpl) GetOrderByID(ctx context.Context, storeID uuid.UUID, orderID uuid.UUID) (*models.Order, error) {
	order, err := s.OrderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.StoreID != storeID {
		return nil, errors.BadRequest("store_id does not match order's store_id")
	}
	return order, nil
}

func (s *OrderServiceImpl) GetOrdersByStoreID(ctx context.Context, storeID uuid.UUID, from, to time.Time) ([]models.Order, error) {
	from, to, err := s.resolveOrderDateRange(from, to)
	if err != nil {
		return nil, err
	}

	return s.OrderRepo.GetByStoreID(
		ctx,
		storeID,
		from,
		to,
	)
}

func (s *OrderServiceImpl) GetDashboardByStoreID(
	ctx context.Context,
	storeID uuid.UUID,
) ([]models.Order, []dtos.OrderSummary, error) {

	orders, err := s.OrderRepo.GetTodayOrders(ctx, storeID)
	if err != nil {
		return nil, nil, err
	}

	summaries, err := s.OrderRepo.GetOrderSummariesByStoreID(ctx, storeID)
	if err != nil {
		return nil, nil, err
	}

	return orders, summaries, nil
}

func (s *OrderServiceImpl) GetSalesReport(ctx context.Context, storeID uuid.UUID, from, to time.Time) (*dtos.SalesReportResponse, error) {
	report, err := s.OrderRepo.GetSalesReport(ctx, storeID, from, to)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s *OrderServiceImpl) GetQueueStoreClient() redis_client.QueueStore {
	return s.RedisQueueStore
}

func (s *OrderServiceImpl) resolveOrderDateRange(
	from time.Time,
	to time.Time,
) (time.Time, time.Time, error) {

	if from.IsZero() && to.IsZero() {
		loc, err := time.LoadLocation("Africa/Lagos")
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		now := time.Now().In(loc)

		from = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0, 0, 0, 0,
			loc,
		)

		to = from.AddDate(0, 0, 1)
	}

	return from, to, nil
}
