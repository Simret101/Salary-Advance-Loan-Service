package domain

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	Role      string    `gorm:"not null" json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Invite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"unique;not null" json:"token"`
	Email     string    `gorm:"not null" json:"email"`
	Expiry    time.Time `gorm:"not null" json:"expiry"`
	Used      bool      `gorm:"default:false" json:"used"`
	InvitedBy uint      `json:"invited_by"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type SendInviteRequest struct {
	Email string `json:"email" binding:"required"`
}


