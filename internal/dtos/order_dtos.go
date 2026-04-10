package dtos

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/models"
)

type CreateOrderRequest struct {
	CustomerName   *string                     `json:"customer_name,omitempty"`
	SoldBy         *string                     `json:"sold_by,omitempty"`
	PaymentMethod  string                      `json:"payment_method" validate:"oneof=CASH TRANSFER"`
	InventoryItems []InventoryOrderItemRequest `json:"inventory_items,omitempty" validate:"dive,required"`
	ComboItems     []ComboOrderItemRequest     `json:"combo_items,omitempty" validate:"dive,required"`
}

type InventoryOrderItemRequest struct {
	InventoryID uuid.UUID `json:"inventory_id" validate:"required,uuid"`
	Quantity    int       `json:"quantity" validate:"required,gt=0"`
}

type ComboOrderItemRequest struct {
	ComboID  uuid.UUID `json:"combo_id" validate:"required,uuid"`
	Quantity int       `json:"quantity" validate:"required,gt=0"`
}

type OrderResponse struct {
	ID            uuid.UUID            `json:"id"`
	CustomerName  *string              `json:"customer_name"`
	TotalAmount   float64              `json:"total_amount"`
	TotalCost     *float64             `json:"total_cost,omitempty"`
	Status        models.OrderStatus   `json:"status"`
	PaymentMethod models.PaymentMethod `json:"payment_method"`
	SoldBy        *string              `json:"sold_by"`
	Channel       models.OrderChannel  `json:"channel"`
	CreatedAt     time.Time            `json:"created_at"`
}

type OrderDetailsResponse struct {
	OrderResponse
	ReceiptNo *string             `json:"receipt_no,omitempty"`
	Items     []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	InventoryID   uuid.UUID  `json:"inventory_id"`
	ProductName   string     `json:"product_name"`
	ImageUrl      *string    `json:"image_url"`
	Quantity      int        `json:"quantity"`
	UnitPrice     float64    `json:"unit_price"`
	UnitCostPrice *float64   `json:"unit_cost_price,omitempty"`
	TotalPrice    float64    `json:"total_price"`
	TotalCost     *float64   `json:"total_cost,omitempty"`
	ComboID       *uuid.UUID `json:"combo_id,omitempty"`
	ComboName     *string    `json:"combo_name,omitempty"`
	QueueNumber   int64     `json:"queue_number,omitempty"`
}

type SalesReportResponse struct {
	StoreID          uuid.UUID          `json:"store_id"`
	From             time.Time          `json:"from"`
	To               time.Time          `json:"to"`
	TotalSales       float64            `json:"total_sales"`
	TotalOrders      int                `json:"total_orders"`
	Profit           float64            `json:"profit"`
	BestSellingItems []BestSellingItem  `json:"best_selling_items"`
	PaymentMethods   map[string]float64 `json:"payment_methods"`
}

type BestSellingItem struct {
	InventoryID  uuid.UUID `json:"inventory_id"`
	ProductName  string    `json:"product_name"`
	QuantitySold int       `json:"quantity_sold"`
	TotalRevenue float64   `json:"total_revenue"`
	Profit       float64   `json:"profit"`
}
