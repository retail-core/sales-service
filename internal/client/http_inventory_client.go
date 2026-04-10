package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/errors"
)

type HttpInventoryClient struct {
	Client     *http.Client
	ServiceURL string
}

func NewHttpInventoryClient(serviceURL string) *HttpInventoryClient {
	return &HttpInventoryClient{
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
		ServiceURL: serviceURL,
	}
}

func (c *HttpInventoryClient) ReserveAndGetSnapshot(ctx context.Context, storeID uuid.UUID, cReq dtos.CreateOrderRequest) (*ReservationResponse, error) {

	reqBody := ReservationRequest{
		InventoryItems: cReq.InventoryItems,
		ComboItems:     cReq.ComboItems,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reservation request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/stores/%s/reservations", c.ServiceURL, storeID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Internal("Failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.ServiceUnavailable("Inventory Service connection failed")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var invError = map[string]interface{}{}
		if err := json.NewDecoder(resp.Body).Decode(&invError); err == nil {
			errMsg := fmt.Sprintf("Reservation Error: %v", invError["message"])
			return nil, errors.BadRequest(errMsg)
		}
		return nil, errors.BadRequest("Inventory Service returned non-OK status")
	}

	var response ReservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode reservation response: %w", err)
	}

	return &ReservationResponse{
		ReservationID:      response.ReservationID,
		InventorySnapshots: response.InventorySnapshots,
		ComboSnapshots:     response.ComboSnapshots,
	}, nil
}
