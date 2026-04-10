package mappers

import (
	"context"
	"fmt"

	"github.com/retail-core/sales-service/internal/dtos"
	"github.com/retail-core/sales-service/internal/models"
	"github.com/retail-core/sales-service/internal/redis_client"
)

func ToOrderResponse(order *models.Order) dtos.OrderResponse {
	return dtos.OrderResponse{
		ID:            order.ID,
		CustomerName:  order.CustomerName,
		TotalAmount:   order.TotalAmount,
		TotalCost:     &order.TotalCost,
		Status:        order.Status,
		PaymentMethod: order.PaymentMethod,
		SoldBy:        order.SoldBy,
		Channel:       order.Channel,
		CreatedAt:     order.CreatedAt,
	}

}

func ToOrderDetailsResponse(order *models.Order, queueStore redis_client.QueueStore, todayOrdersCount *int64) dtos.OrderDetailsResponse {
	items := make([]dtos.OrderItemResponse, len(order.Items))

	var orderNo int64 = 0
	if todayOrdersCount != nil {
		orderNo = *todayOrdersCount + 1
	}

	receiptNo := fmt.Sprintf("%06d", orderNo)


	for i, item := range order.Items {

		ctx := context.Background()
		queueNumber, err := queueStore.GetNextQueueNumber(ctx, item.InventoryID.String())
		if err != nil {
			queueNumber = 0 
		}

		items[i] = dtos.OrderItemResponse{
			InventoryID:   item.InventoryID,
			ProductName:   item.ProductName,
			ImageUrl:      item.ImageUrl,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TotalPrice:    item.Subtotal,
			UnitCostPrice: item.CostPrice,
			TotalCost:     &item.SubtotalCost,
			ComboID:       item.ComboID,
			ComboName:     item.ComboName,
			QueueNumber:   queueNumber,
		}
	}

	return dtos.OrderDetailsResponse{
		OrderResponse: ToOrderResponse(order),
		ReceiptNo:     &receiptNo,
		Items:         items,
	}
}
