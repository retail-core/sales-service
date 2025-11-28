package handler

import (
	"encoding/json"
	"net/http"

	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/service"
)

type OrderHandler struct {
	Service service.OrderService
}

func NewOrderHandler(s service.OrderService) *OrderHandler {
	return &OrderHandler{Service: s}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.Service.Create(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	// Encode the created order (or a simplified response DTO)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		// Log the error but can't change status since headers are sent
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}