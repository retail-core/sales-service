package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

// MockInventoryClient simulates the Inventory Service's reservation logic.
type MockInventoryClient struct{}

// NewMockInventoryClient creates a new mock client.
func NewMockInventoryClient() *MockInventoryClient {
	return &MockInventoryClient{}
}

// ReserveAndGetSnapshot simulates the synchronous reservation and data retrieval.
func (c *MockInventoryClient) ReserveAndGetSnapshot(ctx context.Context, items []ItemRequest) (*ReservationResponse, error) {
	// Simulate connection latency
	time.Sleep(50 * time.Millisecond) 
	
	// --- MOCK DATA ---
	// Price and name (the snapshot data)
	mockInventorySnapshotData := map[string]ProductSnapshot{
		"prod-101": {"prod-101", "Widget X", 49.99},
		"prod-102": {"prod-102", "Gadget Y", 199.99},
		"prod-103": {"prod-103", "A Limited Item", 999.99},
	}
	// Available stock (used for reservation check)
	mockAvailableStock := map[string]int{
		"prod-101": 5,  // Plenty of stock
		"prod-102": 100, // Plenty of stock
		"prod-103": 1,   // Low stock for testing concurrency
	}
	// --- END MOCK DATA ---

	response := &ReservationResponse{
		ReservationID: uuid.Must(uuid.NewV4()), // Generate a mock Reservation ID
		Snapshots:     make([]ProductSnapshot, 0, len(items)),
	}

	for _, item := range items {
		// 1. Check Product Existence & Get Snapshot Data
		snapshot, exists := mockInventorySnapshotData[item.ProductID]
		if !exists {
			// In a real system, you would need to tell the Inventory Service 
			// to roll back any successful reservations made for *other* items in this request.
			return nil, fmt.Errorf("product ID %s not found", item.ProductID) 
		}

		// 2. Check and Simulate Reservation (Concurrency Check)
		available, ok := mockAvailableStock[item.ProductID]
		if !ok || item.Quantity > available {
			// Simulate the critical failure: stock not available
			// In real life, the Inventory DB transaction fails the atomic update.
			return nil, errors.New(fmt.Sprintf("insufficient stock for product %s. Requested: %d, Available: %d", item.ProductID, item.Quantity, available))
		}

		// If successful, add to snapshot list
		response.Snapshots = append(response.Snapshots, snapshot)
	}

	return response, nil
}