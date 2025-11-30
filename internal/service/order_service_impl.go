package service

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/client"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/models"
	"github.com/retail-core/sales-service/internal/mq"
	"github.com/retail-core/sales-service/internal/repository"
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

func (s *OrderServiceImpl) Create(ctx context.Context, req dtos.CreateOrderRequest) (*models.Order, error) {
	// 1. **Inventory Reservation & Snapshot (Simulation)**
	// In a real scenario, this is where we call the Inventory Service:
	// 		inventoryData, reservationID, err := s.InventoryClient.ReserveAndGetSnapshot(ctx, req.Items)
	// For now, we'll use mock data and assume the reservation was successful.
	
	// --- MOCK INVENTORY RESPONSE START ---
	// Assume this is the data returned from a successful Inventory Service call 
	// that simultaneously reserved the stock.
	mockInventorySnapshot := map[string]struct{ Name string; Price float64; Available int }{
		"prod-101": {"Widget X", 49.99, 5},
		"prod-102": {"Gadget Y", 199.99, 100},
	}
	// reservationID := uuid.Must(uuid.NewV4()) // Mock Reservation ID
	// --- MOCK INVENTORY RESPONSE END ---

	// 2. **Build Order and Items (Snapshot & Calculation)**
	newOrder := &models.Order{
		Base:      		models.Base{ID: uuid.Must(uuid.NewV4())},
		CustomerName: 	&req.CustomerID,
		Status:    "COMPLETED", // Start in PENDING until stock deduction is confirmed
		TotalAmount: 0.0,
		Items:     make([]models.OrderItem, 0, len(req.Items)),
	}
	
	var calculatedTotal float64

	for _, itemReq := range req.Items {
		snapshot, exists := mockInventorySnapshot[itemReq.InventoryID]
		
		if !exists {
			// In real life, we would rollback the reservation and return an error.
			return nil, fmt.Errorf("inventory ID %s not found in inventory", itemReq.InventoryID)
		}
		
		if itemReq.Quantity <= 0 || itemReq.Quantity > snapshot.Available {
			// In real life, we would rollback the reservation and return an error.
			return nil, fmt.Errorf("invalid quantity or insufficient stock for inventory ID %s. Available: %d, Requested: %d", itemReq.InventoryID, snapshot.Available, itemReq.Quantity)
		}

		subtotal := float64(itemReq.Quantity) * snapshot.Price
		calculatedTotal += subtotal

		orderItem := models.OrderItem{
			Base:                  models.Base{ID: uuid.Must(uuid.NewV4())},
			InventoryID:           itemReq.InventoryID,
			ProductName:   		   snapshot.Name,
			UnitPrice:     		   snapshot.Price,
			Quantity:              itemReq.Quantity,
			Subtotal:              subtotal,
		}
		newOrder.Items = append(newOrder.Items, orderItem)
	}

	newOrder.TotalAmount = calculatedTotal

	// 3. **Persistence (Repository)**
	savedOrder, err := s.OrderRepo.CreateOrder(ctx, newOrder)
	if err != nil {
		// In a failure scenario here, we MUST publish a RabbitMQ event 
		// to the Inventory Service to ROLLBACK the reservation!
		// s.MQPublisher.PublishRollback(reservationID)
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// 4. **Asynchronous Stock Confirmation (RabbitMQ)**
	// Order saved successfully, so we publish an event to confirm the deduction.
	// s.MQPublisher.PublishConfirmation(savedOrder.ID, reservationID)

	return savedOrder, nil
}

func (s *OrderServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	return s.OrderRepo.FindByID(ctx, id)
}

func (s *OrderServiceImpl) UpdateStatusByEvent(ctx context.Context, orderID uuid.UUID, newStatus string) error {
	// Simple validation to ensure the status is a valid change.
	if newStatus == "COMPLETED" || newStatus == "STOCK_ERROR" {
		return s.OrderRepo.UpdateStatus(ctx, orderID, newStatus)
	}
	return fmt.Errorf("invalid status provided for update: %s", newStatus)
}