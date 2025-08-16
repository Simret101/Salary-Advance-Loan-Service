package usecases

import (
	"context"
	"fmt"
	"time"
	"SalaryAdvance/internal/domain"
	
	"SalaryAdvance/pkg/config"
)

type InviteUseCaseImpl struct {
	inviteRepo   domain.InviteRepository
	userRepo     domain.UserRepository
	emailService domain.EmailService
	jwtService   domain.JWTService
}

func NewInviteUseCase(inviteRepo domain.InviteRepository, userRepo domain.UserRepository, emailService domain.EmailService, jwtService domain.JWTService) *InviteUseCaseImpl {
	return &InviteUseCaseImpl{inviteRepo: inviteRepo, userRepo: userRepo, emailService: emailService, jwtService: jwtService}
}

func (u *InviteUseCaseImpl) SendInvite(ctx context.Context, adminID uint, email string) (string, error) {
	if _, err := u.userRepo.FindByEmail(ctx, email); err == nil {
		return "", config.ErrBadRequest
	}
	inviteToken, err := u.jwtService.GenerateInviteToken(email, adminID)
	if err != nil {
		return "", config.ErrInternalServer
	}
	invite := &domain.Invite{
		Token:     inviteToken,
		Email:     email,
		Expiry:    time.Now().Add(24 * time.Hour),
		InvitedBy: adminID,
	}
	_, err = u.inviteRepo.Create(ctx, invite)
	if err != nil {
		return "", config.ErrInternalServer
	}
	link := fmt.Sprintf("http://localhost:8080/register?token=%s", inviteToken)
	if err := u.emailService.SendInvite(email, link); err != nil {
		fmt.Println("Email send error:", err)
	}
	return link, nil
}

func (u *InviteUseCaseImpl) ValidateInvite(ctx context.Context, token string) (string, error) {
	invite, err := u.inviteRepo.FindByToken(ctx, token)
	if err != nil {
		return "", config.ErrNotFound
	}
	if invite.Used || invite.Expiry.Before(time.Now()) {
		return "", config.ErrBadRequest
	}
	_, err = u.jwtService.ValidateToken(token)
	if err != nil {
		return "", config.ErrUnauthorized
	}
	invite.Used = true
	_, err = u.inviteRepo.Update(ctx, invite)
	if err != nil {
		return "", config.ErrInternalServer
	}
	return invite.Email, nil
}