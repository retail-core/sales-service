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
	Quantity    int32   `json:"quantity,omitempty"`
	ImageUrl    string  `json:"image_url,omitempty"`
}

type ReservationResponse struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	Snapshots     []InventorySnapshot `json:"snapshots"`
}

type ReservationRequest struct {
	Items []dtos.OrderItemRequest `json:"items" validate:"required,dive,required"`
}

type InventoryClient interface {
	ReserveAndGetSnapshot(ctx context.Context, storeID uuid.UUID, items []dtos.OrderItemRequest) (*ReservationResponse, error)
}
