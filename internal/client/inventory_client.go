package client

import (
	"context"

	"github.com/gofrs/uuid"
)

// ItemRequest represents the minimal data sent for reservation/lookup.
type ItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// ProductSnapshot represents the critical data returned by the Inventory Service.
type ProductSnapshot struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"unit_price"`
	// Note: We don't need 'AvailableStock' back, as the reservation step handles availability.
}

// ReservationResponse bundles the snapshots and the reservation ID.
type ReservationResponse struct {
	ReservationID uuid.UUID
	Snapshots     []ProductSnapshot
}

// InventoryClient defines the contract for synchronous interaction with the Inventory Service.
type InventoryClient interface {
	// ReserveAndGetSnapshot attempts to claim stock and retrieve immutable details 
	// for the requested items. Returns a Reservation ID on success.
	ReserveAndGetSnapshot(ctx context.Context, items []ItemRequest) (*ReservationResponse, error)
}