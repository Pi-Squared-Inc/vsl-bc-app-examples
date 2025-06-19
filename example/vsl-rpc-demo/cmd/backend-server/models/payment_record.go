package models

type UserPaymentRecord struct {
	ID          string `gorm:"primaryKey"`
	UserAddress string `gorm:"column:user_address"`
}
