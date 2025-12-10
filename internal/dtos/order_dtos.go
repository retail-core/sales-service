package dtos

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/models"
)

type CreateOrderRequest struct {
	CustomerName  *string            `json:"customer_name,omitempty"`
	SoldBy        *string            `json:"sold_by,omitempty"`
	PaymentMethod string             `json:"payment_method" validate:"oneof=CASH TRANSFER"`
	Items         []OrderItemRequest `json:"items" validate:"required,min=1,dive,required"`
}

type OrderItemRequest struct {
	InventoryID uuid.UUID `json:"inventory_id" validate:"required,uuid"`
	Quantity    int       `json:"quantity" validate:"required,gt=0"`
}

type OrderResponse struct {
	ID            uuid.UUID            `json:"id"`
	CustomerName  *string              `json:"customer_name"`
	TotalAmount   float64              `json:"total_amount"`
	Status        models.OrderStatus   `json:"status"`
	PaymentMethod models.PaymentMethod `json:"payment_method"`
	SoldBy        *string              `json:"sold_by"`
	Channel       models.OrderChannel  `json:"channel"`
	CreatedAt     time.Time            `json:"created_at"`
}

type OrderDetailsResponse struct {
	OrderResponse
	Items []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	InventoryID uuid.UUID `json:"inventory_id"`
	ProductName  string    `json:"product_name"`
	ImageUrl	*string   `json:"image_url"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TotalPrice  float64   `json:"total_price"`
}