package dtos

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/models"
)

type OrderSummaryType string

const (
	OrderSummaryYesterday OrderSummaryType = "yesterday"
	OrderSummaryMonth     OrderSummaryType = "month"
)

type OrderSummary struct {
	Type            OrderSummaryType `json:"type"`
	Month           int              `json:"month"`
	Year            int              `json:"year"`
	From            time.Time        `json:"from"`
	To              time.Time        `json:"to"`
	TotalOrderCount int64            `json:"total_order_count"`
	TotalRevenue    float64          `json:"total_revenue"`
	TotalProfit     float64          `json:"total_profit"`
}

type GetOrdersByStoreResponse struct {
	Orders    []OrderResponse `json:"orders"`
	Summaries []OrderSummary  `json:"summaries"`
}

type CreateOrderRequest struct {
	CustomerName   *string                     `json:"customer_name,omitempty"`
	SoldBy         *string                     `json:"sold_by,omitempty"`
	PaymentMethod  string                      `json:"payment_method" validate:"oneof=CASH TRANSFER"`
	InventoryItems []InventoryOrderItemRequest `json:"inventory_items,omitempty" validate:"dive,required"`
}

type InventoryOrderItemRequest struct {
	InventoryID uuid.UUID `json:"inventory_id" validate:"required,uuid"`
	UnitID      uuid.UUID `json:"unit_id" validate:"required,uuid"`
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
	InventoryID   uuid.UUID `json:"inventory_id"`
	ProductName   string    `json:"product_name"`
	ImageUrl      *string   `json:"image_url"`
	Quantity      int       `json:"quantity"`
	UnitPrice     float64   `json:"unit_price"`
	UnitCostPrice *float64  `json:"unit_cost_price,omitempty"`
	TotalPrice    float64   `json:"total_price"`
	TotalCost     *float64  `json:"total_cost,omitempty"`
	QueueNumber   int64     `json:"queue_number,omitempty"`

	UnitLabel  string `json:"unit_label"`
	QtyPerUnit int    `json:"qty_per_unit"`
	IsBaseUnit bool   `json:"is_base_unit"`
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
