package domain

import "time"

type Transaction struct {
	ID                  int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerID          string    `gorm:"type:varchar(255);not null" json:"customer_id"`
	FromAccount         string    `gorm:"type:varchar(255)" json:"fromAccount"`
	ToAccount           string    `gorm:"type:varchar(255)" json:"toAccount"`
	Amount              float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Remark              string    `gorm:"type:varchar(255)" json:"remark"`
	TransactionType     string    `gorm:"type:varchar(50);not null" json:"transactionType"` 
	RequestId           string    `gorm:"type:varchar(255)" json:"requestId"`
	Reference           string    `gorm:"type:varchar(255)" json:"reference"`
	ThirdPartyReference string    `gorm:"type:varchar(255)" json:"thirdPartyReference"`
	InstitutionId       string    `gorm:"type:varchar(255)" json:"institutionId"`
	ClearedBalance      float64   `gorm:"type:decimal(15,2)" json:"clearedBalance"`
	TransactionDate     time.Time `gorm:"type:timestamp;not null" json:"transactionDate"`
	BillerId            string    `gorm:"type:varchar(255)" json:"billerId"`
	Status              string    `gorm:"type:varchar(50);not null" json:"status"` 
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Customer            Customer  `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}
