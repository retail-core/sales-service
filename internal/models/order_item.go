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

	CostPrice   *float64 `gorm:"type:numeric(10, 2);"`
	
	Quantity int `gorm:"not null"` 
	
	Subtotal float64 `gorm:"type:numeric(10, 2);not null"`
	SubtotalCost float64 `gorm:"type:numeric(10, 2);not null;default:0"`

	// Selling unit snapshot — captured at time of sale so historical
	// orders stay accurate even if the inventory's units change later.
	UnitLabel  string    `gorm:"type:varchar(100);not null;default:'Piece'"`
	QtyPerUnit int       `gorm:"not null;default:1"`
	IsBaseUnit *bool      `gorm:"not null;default:true"`
	
}

func (OrderItem) TableName() string {
	return "order_items"
}