package handler

import (
	"encoding/json"
	"net/http"

	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/httpx"
	"github.com/retail-core/sales-service/internal/service"
	"github.com/retail-core/sales-service/internal/validation"
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
		httpx.WriteError(w, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	order, err := h.Service.Create(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, order)
}