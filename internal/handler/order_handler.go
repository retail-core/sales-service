package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/errors"
	"github.com/retail-core/sales-service/internal/httpx"
	"github.com/retail-core/sales-service/internal/mappers"
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
	storeIDParam := chi.URLParam(r, "store_id")

	storeID, err := uuid.FromString(storeIDParam)
	if err != nil {
		httpx.WriteError(w, errors.BadRequest("store_id must be valid UUID string"))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errors.BadRequest("Invalid Request Body"))
		return
	}

	if err := validation.ValidateStruct(req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	order, err := h.Service.Create(r.Context(), storeID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, mappers.ToOrderResponse(order))
}