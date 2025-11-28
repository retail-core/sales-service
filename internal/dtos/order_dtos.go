package dtos

type CreateOrderRequest struct {
	CustomerID string `json:"customer_id"`
	Items []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	InventoryID string `json:"inventory_id"`
	Quantity    int    `json:"quantity"`
}
