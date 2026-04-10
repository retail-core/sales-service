package handler

import (
	"encoding/json"
	"net/http"
	"time"

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

	order, todayOrdersCount, err := h.Service.Create(r.Context(), storeID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	queueStoreClient := h.Service.GetQueueStoreClient()

	httpx.WriteJSON(w, http.StatusCreated, mappers.ToOrderDetailsResponse(order, queueStoreClient, &todayOrdersCount))
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	storeIDParam := chi.URLParam(r, "store_id")
	orderIDParam := chi.URLParam(r, "order_id")

	storeID, err := uuid.FromString(storeIDParam)
	if err != nil {
		httpx.WriteError(w, errors.BadRequest("store_id must be valid UUID string"))
		return
	}

	orderID, err := uuid.FromString(orderIDParam)
	if err != nil {
		httpx.WriteError(w, errors.BadRequest("order_id must be valid UUID string"))
		return
	}

	order, err := h.Service.GetOrderByID(r.Context(), storeID, orderID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, mappers.ToOrderDetailsResponse(order, h.Service.GetQueueStoreClient(), nil))
}

func (h *OrderHandler) GetOrdersByStoreID(w http.ResponseWriter, r *http.Request) {
	storeIDParam := chi.URLParam(r, "store_id")

	storeID, err := uuid.FromString(storeIDParam)
	if err != nil {
		httpx.WriteError(w, errors.BadRequest("store_id must be valid UUID string"))
		return
	}

	orders, err := h.Service.GetOrdersByStoreID(r.Context(), storeID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	response := make([]dtos.OrderResponse, len(orders))

	for i, order := range orders {
		response[i] = mappers.ToOrderResponse(&order)
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *OrderHandler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	storeIDParam := chi.URLParam(r, "store_id")

	storeID, err := uuid.FromString(storeIDParam)
	if err != nil {
		httpx.WriteError(w, errors.BadRequest("store_id must be valid UUID string"))
		return
	}

	from, to, err := parseDateRange(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	report, err := h.Service.GetSalesReport(r.Context(), storeID, from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, report)
}

func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	now := time.Now()
	loc := now.Location()

	// default: today
	if fromStr == "" && toStr == "" {
		from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		to := from.Add(24 * time.Hour)
		return from, to, nil
	}

	from, err := time.ParseInLocation("2006-01-02", fromStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, errors.BadRequest("from must be YYYY-MM-DD")
	}

	to, err := time.ParseInLocation("2006-01-02", toStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, errors.BadRequest("to must be YYYY-MM-DD")
	}

	to = to.Add(24 * time.Hour) // inclusive

	if to.Before(from) {
		return time.Time{}, time.Time{}, errors.BadRequest("to must be after from")
	}

	return from, to, nil
}
