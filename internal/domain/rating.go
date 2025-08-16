package domain

import (
	"time"
)

type Rating struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerID string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"customer_id"`
	Score      float64   `gorm:"type:decimal(5,2);not null" json:"score"`
	Breakdown  struct {
		CountScore     float64 `json:"count_score"`
		VolumeScore    float64 `json:"volume_score"`
		DurationScore  float64 `json:"duration_score"`
		StabilityScore float64 `json:"stability_score"`
	} `gorm:"type:jsonb" json:"breakdown"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Customer  Customer  `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}
