package usecases

import (
	"context"
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/services"
	"SalaryAdvance/pkg/config"
)

type AuthUseCaseImpl struct {
	userRepo   domain.UserRepository
	jwtService domain.JWTService
}

func NewAuthUseCase(userRepo domain.UserRepository, jwtService domain.JWTService) *AuthUseCaseImpl {
	return &AuthUseCaseImpl{userRepo: userRepo, jwtService: jwtService}
}

func (u *AuthUseCaseImpl) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", config.ErrNotFound
	}
	if !services.CheckPasswordHash(password, user.Password) {
		return "", "", config.ErrUnauthorized
	}
	accessToken, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", config.ErrInternalServer
	}
	refreshToken, err := u.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", config.ErrInternalServer
	}
	return accessToken, refreshToken, nil
}

func (u *AuthUseCaseImpl) RegisterFromInvite(ctx context.Context, token, password string) error {
	claims, err := u.jwtService.ValidateToken(token)
	if err != nil {
		return config.ErrUnauthorized
	}
	email, ok := claims["email"].(string)
	if !ok {
		return config.ErrBadRequest
	}
	hashedPassword, err := services.HashPassword(password)
	if err != nil {
		return config.ErrInternalServer
	}
	user := &domain.User{
		Email:    email,
		Password: hashedPassword,
		Role:     "uploader",
	}
	_, err = u.userRepo.Create(ctx, user)
	if err != nil {
		return config.ErrInternalServer
	}
	return nil
}

func (u *AuthUseCaseImpl) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := u.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return "", config.ErrUnauthorized
	}
	userID, ok := claims["id"].(float64)
	if !ok {
		return "", config.ErrBadRequest
	}
	user := &domain.User{ID: uint(userID)}
	accessToken, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", config.ErrInternalServer
	}
	return accessToken, nil
}