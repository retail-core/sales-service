package models

import (
	"github.com/gofrs/uuid"
)

type OrderItem struct {
	Base

	OrderID uuid.UUID `gorm:"index;not null"` 

	InventoryID uuid.UUID `gorm:"type:varchar(100);not null"`
	
	ProductName string `gorm:"type:varchar(255);not null"`

	ImageUrl *string `gorm:"type:text;"`
	
	UnitPrice   float64 `gorm:"type:numeric(10, 2);not null"` 
	
	Quantity int `gorm:"not null"` 
	
	Subtotal float64 `gorm:"type:numeric(10, 2);not null"` 
}

func (OrderItem) TableName() string {
	return "order_items"
}