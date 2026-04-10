package client

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/dtos"
)

type InventorySnapshot struct {
	InventoryID string  `json:"inventory_id"`
	Name        string  `json:"name"`
	UnitPrice   float64 `json:"unit_price"`
	CostPrice   *float64 `json:"cost_price,omitempty"`
	Quantity    int32   `json:"quantity,omitempty"`
	ImageUrl    string  `json:"image_url,omitempty"`
}

type ComboSnapshot struct {
	ComboID   string  `json:"combo_id"`
	Name      string  `json:"name"`
	UnitPrice float64 `json:"unit_price"`
	Items    []InventorySnapshot `json:"items"`
}

type ReservationResponse struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	InventorySnapshots     []InventorySnapshot `json:"inventory_snapshots,omitempty"`
	ComboSnapshots         []ComboSnapshot `json:"combo_snapshots,omitempty"`
}

type ReservationRequest struct {
	InventoryItems []dtos.InventoryOrderItemRequest `json:"inventory_items" validate:"required,dive,required"`
	ComboItems     []dtos.ComboOrderItemRequest     `json:"combo_items,omitempty" validate:"dive,required"`
}

type InventoryClient interface {
	ReserveAndGetSnapshot(ctx context.Context, storeID uuid.UUID, req dtos.CreateOrderRequest) (*ReservationResponse, error)
}
