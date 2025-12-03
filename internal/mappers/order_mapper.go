package mappers

import (
	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/models"
)

func ToOrderResponse(order *models.Order) dtos.OrderResponse {
	return dtos.OrderResponse{
		ID:            order.ID,
		CustomerName:  order.CustomerName,
		TotalAmount:   order.TotalAmount,
		Status:        order.Status,
		PaymentMethod: order.PaymentMethod,
		SoldBy:        order.SoldBy,
		Channel:       order.Channel,
		CreatedAt:    order.CreatedAt,
	}

}