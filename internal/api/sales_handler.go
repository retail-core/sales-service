package api

import (
	"net/http"
	"github.com/retail-core/sales-service/internal/sales"
)

type SalesHandler struct {
	service sales.Service
}

func NewSalesHandler(service sales.Service) *SalesHandler {
	return &SalesHandler{service: service}
}

func (h *SalesHandler) makeSaleHandler(w http.ResponseWriter, r *http.Request) {
}