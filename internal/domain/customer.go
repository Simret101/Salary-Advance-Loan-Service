package domain

import (
	"encoding/json"
	"fmt"
	"time"
)


type AccountNo string


type Customer struct {
	ID              int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerId      string    `gorm:"type:varchar(255);unique;not null" validate:"required" json:"customerId"`
	CustomerName    string    `gorm:"type:varchar(255);not null" validate:"required,min=3,max=255" json:"customerName"`
	Mobile          string    `gorm:"type:varchar(255)" validate:"omitempty,min=10,max=15" json:"mobile"`
	AccountNo       AccountNo `gorm:"type:varchar(255);not null" validate:"required" json:"accountNo"`
	BranchName      string    `gorm:"type:varchar(255)" json:"branchName"`
	BranchCode      string    `gorm:"type:varchar(255)" json:"branchCode"`
	ProductName     string    `gorm:"type:varchar(255)" json:"productName"`
	CustomerBalance float64   `gorm:"type:decimal(15,2);default:0" json:"customerBalance"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
func (a *AccountNo) UnmarshalJSON(data []byte) error {
	var num float64
	var str string

	
	if err := json.Unmarshal(data, &num); err == nil {
		*a = AccountNo(fmt.Sprintf("%.0f", num))
		return nil
	}

	
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("account_no must be a number or string, got %s", string(data))
	}

	*a = AccountNo(str)
	return nil
}