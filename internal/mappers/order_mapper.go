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
		CreatedAt:     order.CreatedAt,
	}

}

func ToOrderDetailsResponse(order *models.Order) dtos.OrderDetailsResponse {
	items := make([]dtos.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = dtos.OrderItemResponse{
			InventoryID: item.InventoryID,
			ProductName: item.ProductName,
			ImageUrl:    item.ImageUrl,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.Subtotal,
		}
	}

	return dtos.OrderDetailsResponse{
		OrderResponse: ToOrderResponse(order),
		Items:         items,
	}
}
