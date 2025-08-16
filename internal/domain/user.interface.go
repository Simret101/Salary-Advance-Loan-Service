package domain

import (
	"context"
)


type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
}


type InviteRepository interface {
	Create(ctx context.Context, invite *Invite) (*Invite, error)
	FindByToken(ctx context.Context, token string) (*Invite, error)
	Update(ctx context.Context, invite *Invite) (*Invite, error)
}


type JWTService interface {
	GenerateAccessToken(user *User) (string, error)
	GenerateRefreshToken(user *User) (string, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
	GenerateInviteToken(email string, invitedBy uint) (string, error)
}


type EmailService interface {
	SendInvite(email, link string) error
}


type AuthUseCase interface {
	Login(ctx context.Context, email, password string) (string, string, error)
	RegisterFromInvite(ctx context.Context, token, password string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
}


type InviteUseCase interface {
	SendInvite(ctx context.Context, adminID uint, email string) (string, error)
	ValidateInvite(ctx context.Context, token string) (string, error)
}