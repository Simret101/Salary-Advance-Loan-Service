package domain

import "time"

type Transaction struct {
	TransactionID string    `gorm:"type:varchar(255);unique;not null" validate:"required" json:"transactionId"`
	FromAccount   AccountNo `gorm:"type:varchar(255);not null" validate:"required" json:"fromAccount"`
	ToAccount     AccountNo `gorm:"type:varchar(255);not null" validate:"required" json:"toAccount"`
	Amount        float64   `gorm:"type:decimal(15,2);not null" validate:"required,gt=0" json:"amount"`
	Date          time.Time `gorm:"type:date;not null" validate:"required" json:"date"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
