package service

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/client"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/errors"
	"github.com/retail-core/sales-service/internal/logger"
	"github.com/retail-core/sales-service/internal/models"
	"github.com/retail-core/sales-service/internal/mq"
	"github.com/retail-core/sales-service/internal/repository"
	"go.uber.org/zap"
	// TODO: Inventory service client import will go here (e.g., github.com/your-username/inventory-client)
	// TODO: RabbitMQ client import will go here (e.g., github.com/streadway/amqp)
)

// Implementation Note: In a real microservice, we would inject interfaces for
// the InventoryClient and MessageQueuePublisher here as well.

// OrderServiceImpl is the concrete implementation of OrderService.
type OrderServiceImpl struct {
	OrderRepo repository.OrderRepository
	InventoryClient client.InventoryClient
	MQPublisher mq.MessageQueuePublisher
}

func NewOrderServiceImpl(repo repository.OrderRepository, inventoryClient client.InventoryClient, mqPublisher mq.MessageQueuePublisher) *OrderServiceImpl {
	return &OrderServiceImpl{
		OrderRepo: repo,
		InventoryClient: inventoryClient,
		MQPublisher: mqPublisher,
	}
}

func (s *OrderServiceImpl) Create(ctx context.Context, storeID uuid.UUID, req dtos.CreateOrderRequest) (*models.Order, error) {

	res, err := s.InventoryClient.ReserveAndGetSnapshot(ctx, storeID, req.Items)
	if err != nil {
		return nil, err
	}

	snapshotMap := make(map[string]client.InventorySnapshot)
	for _, snapshot := range res.Snapshots {
		snapshotMap[snapshot.InventoryID] = snapshot
	}

	orderID := uuid.Must(uuid.NewV4())
	newOrder := &models.Order{
		Base:      		models.Base{ID: orderID},
		StoreID:        storeID,
		CustomerName: 	req.CustomerName,
		SoldBy:         req.SoldBy,
		PaymentMethod:  models.PaymentMethod(req.PaymentMethod),
		Channel:        models.OrderChannelInStore,
		Status:    		models.OrderCompleted, 
		TotalAmount: 0.0,
		Items:     make([]models.OrderItem, 0, len(req.Items)),
	}

	var calculatedTotal float64
	for _, itemReq := range req.Items {
		snapshot, exists := snapshotMap[itemReq.InventoryID.String()]
		
		if !exists {
			s.MQPublisher.PublishRollback(ctx, res.ReservationID)
			return nil, fmt.Errorf("inventory snapshot not found for inventory ID: %s", itemReq.InventoryID)
		}

		subtotal := float64(itemReq.Quantity) * snapshot.UnitPrice
		calculatedTotal += subtotal

		orderItem := models.OrderItem{
			Base:                  models.Base{ID: uuid.Must(uuid.NewV4())},
			InventoryID:           itemReq.InventoryID,
			OrderID:               orderID,
			ProductName:   		   snapshot.Name,
			ImageUrl:              &snapshot.ImageUrl,
			UnitPrice:     		   snapshot.UnitPrice,
			Quantity:              itemReq.Quantity,
			Subtotal:              subtotal,
		}
		newOrder.Items = append(newOrder.Items, orderItem)
	}

	newOrder.TotalAmount = calculatedTotal

	savedOrder, err := s.OrderRepo.CreateOrder(ctx, newOrder)
	if err != nil {
		s.MQPublisher.PublishRollback(ctx, res.ReservationID)
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	if err := s.MQPublisher.PublishConfirmation(ctx, res.ReservationID); err != nil {
		// Retry Logic
		logger.L().Warn("WARNING: failed to publish confirmation event", zap.Error(err))
	}

	return savedOrder, nil
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

func (s *OrderServiceImpl) GetOrdersByStoreID(ctx context.Context, storeID uuid.UUID) ([]models.Order, error) {
	orders, err := s.OrderRepo.GetByStoreID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}