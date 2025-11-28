package models

type OrderStatus string
type PaymentMethod string
type OrderChannel string

const (
    OrderPending   OrderStatus = "PENDING"
    OrderCompleted OrderStatus = "COMPLETED"
    OrderCanceled  OrderStatus = "CANCELED"
)

const (
	PaymentCash PaymentMethod = "CASH"
	PaymentBankTransfer PaymentMethod = "TRANSFER"
)

const (
	OrderChannelOnline  OrderChannel = "ONLINE"
	OrderChannelInStore OrderChannel = "IN_STORE"
)


type Order struct {
	Base

	CustomerName 	*string			`gorm:"type:varchar(255);default:null"`

	TotalAmount 	float64 		`gorm:"type:numeric(10, 2);not null"`

	Status       	OrderStatus 	`gorm:"type:enum('PENDING','COMPLETED','CANCELED');default:'PENDING'"`

	PaymentMethod 	PaymentMethod 	`gorm:"type:enum('CASH','TRANSFER');default:'CASH'"`

	Channel      	OrderChannel 	`gorm:"type:enum('ONLINE','IN_STORE');default:'IN_STORE'"`

	SoldBy	  		*string 		`gorm:"type:varchar(100);default:null"`
	
	Items	   		[]OrderItem   	`gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;"`

}

func (Order) TableName() string {
	return "orders"
}

