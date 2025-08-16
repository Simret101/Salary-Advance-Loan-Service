package usecases

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/mocks"
	"SalaryAdvance/pkg/config"
	"context"
	"errors"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendInvite(email, link string) error {
	args := m.Called(email, link)
	return args.Error(0)
}

func TestInviteUseCase_SendInvite(t *testing.T) {
	ctx := context.Background()
	mockInviteRepo := mocks.NewInviteRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)
	mockEmailService := &MockEmailService{}
	mockJWTService := &MockJWTService{}
	uc := NewInviteUseCase(mockInviteRepo, mockUserRepo, mockEmailService, mockJWTService)

	tests := []struct {
		name         string
		adminID      uint
		email        string
		mockSetup    func()
		expectedLink string
		expectedErr  error
	}{
		{
			name:    "Valid invite",
			adminID: 1,
			email:   "test@example.com",
			mockSetup: func() {
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(nil, errors.New("not found")).Once()
				mockJWTService.On("GenerateInviteToken", "test@example.com", uint(1)).Return("invite_token", nil).Once()
				invite := &domain.Invite{
					Token:     "invite_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(24 * time.Hour),
					InvitedBy: 1,
				}
				mockInviteRepo.On("Create", ctx, invite).Return(invite, nil).Once()
				mockEmailService.On("SendInvite", "test@example.com", "http://localhost:8080/register?token=invite_token").Return(nil).Once()
			},
			expectedLink: "http://localhost:8080/register?token=invite_token",
			expectedErr:  nil,
		},
		{
			name:    "User already exists",
			adminID: 1,
			email:   "test@example.com",
			mockSetup: func() {
				user := &domain.User{ID: 1, Email: "test@example.com"}
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil).Once()
			},
			expectedLink: "",
			expectedErr:  config.ErrBadRequest,
		},
		{
			name:    "Token generation failure",
			adminID: 1,
			email:   "test@example.com",
			mockSetup: func() {
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(nil, errors.New("not found")).Once()
				mockJWTService.On("GenerateInviteToken", "test@example.com", uint(1)).Return("", errors.New("token error")).Once()
			},
			expectedLink: "",
			expectedErr:  config.ErrInternalServer,
		},
		{
			name:    "Invite creation failure",
			adminID: 1,
			email:   "test@example.com",
			mockSetup: func() {
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(nil, errors.New("not found")).Once()
				mockJWTService.On("GenerateInviteToken", "test@example.com", uint(1)).Return("invite_token", nil).Once()
				invite := &domain.Invite{
					Token:     "invite_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(24 * time.Hour),
					InvitedBy: 1,
				}
				mockInviteRepo.On("Create", ctx, invite).Return(nil, errors.New("create error")).Once()
			},
			expectedLink: "",
			expectedErr:  config.ErrInternalServer,
		},
		{
			name:    "Email send failure",
			adminID: 1,
			email:   "test@example.com",
			mockSetup: func() {
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(nil, errors.New("not found")).Once()
				mockJWTService.On("GenerateInviteToken", "test@example.com", uint(1)).Return("invite_token", nil).Once()
				invite := &domain.Invite{
					Token:     "invite_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(24 * time.Hour),
					InvitedBy: 1,
				}
				mockInviteRepo.On("Create", ctx, invite).Return(invite, nil).Once()
				mockEmailService.On("SendInvite", "test@example.com", "http://localhost:8080/register?token=invite_token").Return(errors.New("email error")).Once()
			},
			expectedLink: "http://localhost:8080/register?token=invite_token",
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			link, err := uc.SendInvite(ctx, tt.adminID, tt.email)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Equal(t, tt.expectedLink, link)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLink, link)
			}
		})
	}
}

func TestInviteUseCase_ValidateInvite(t *testing.T) {
	ctx := context.Background()
	mockInviteRepo := mocks.NewInviteRepository(t)
	mockUserRepo := mocks.NewUserRepository(t)
	mockEmailService := &MockEmailService{}
	mockJWTService := &MockJWTService{}
	uc := NewInviteUseCase(mockInviteRepo, mockUserRepo, mockEmailService, mockJWTService)

	tests := []struct {
		name          string
		token         string
		mockSetup     func()
		expectedEmail string
		expectedErr   error
	}{
		{
			name:  "Valid invite",
			token: "valid_token",
			mockSetup: func() {
				invite := &domain.Invite{
					Token:     "valid_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(1 * time.Hour),
					Used:      false,
					InvitedBy: 1,
				}
				mockInviteRepo.On("FindByToken", ctx, "valid_token").Return(invite, nil).Once()
				mockJWTService.On("ValidateToken", "valid_token").Return(map[string]interface{}{"email": "test@example.com"}, nil).Once()
				updatedInvite := *invite 
				updatedInvite.Used = true
				mockInviteRepo.On("Update", ctx, &updatedInvite).Return(&updatedInvite, nil).Once()
			},
			expectedEmail: "test@example.com",
			expectedErr:   nil,
		},
		{
			name:  "Invite not found",
			token: "invalid_token",
			mockSetup: func() {
				mockInviteRepo.On("FindByToken", ctx, "invalid_token").Return(nil, errors.New("not found")).Once()
			},
			expectedEmail: "",
			expectedErr:   config.ErrNotFound,
		},
		{
			name:  "Invite already used",
			token: "used_token",
			mockSetup: func() {
				invite := &domain.Invite{
					Token:     "used_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(1 * time.Hour),
					Used:      true,
					InvitedBy: 1,
				}
				mockInviteRepo.On("FindByToken", ctx, "used_token").Return(invite, nil).Once()
			},
			expectedEmail: "",
			expectedErr:   config.ErrBadRequest,
		},
		{
			name:  "Invite expired",
			token: "expired_token",
			mockSetup: func() {
				invite := &domain.Invite{
					Token:     "expired_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(-1 * time.Hour),
					Used:      false,
					InvitedBy: 1,
				}
				mockInviteRepo.On("FindByToken", ctx, "expired_token").Return(invite, nil).Once()
			},
			expectedEmail: "",
			expectedErr:   config.ErrBadRequest,
		},
		{
			name:  "Invalid token",
			token: "invalid_token",
			mockSetup: func() {
				invite := &domain.Invite{
					Token:     "invalid_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(1 * time.Hour),
					Used:      false,
					InvitedBy: 1,
				}
				mockInviteRepo.On("FindByToken", ctx, "invalid_token").Return(invite, nil).Once()
				mockJWTService.On("ValidateToken", "invalid_token").Return(map[string]interface{}{}, errors.New("invalid token")).Once()
			},
			expectedEmail: "",
			expectedErr:   config.ErrUnauthorized,
		},
		{
			name:  "Update failure",
			token: "valid_token",
			mockSetup: func() {
				invite := &domain.Invite{
					Token:     "valid_token",
					Email:     "test@example.com",
					Expiry:    time.Now().Add(1 * time.Hour),
					Used:      false,
					InvitedBy: 1,
				}
				mockInviteRepo.On("FindByToken", ctx, "valid_token").Return(invite, nil).Once()
				mockJWTService.On("ValidateToken", "valid_token").Return(map[string]interface{}{"email": "test@example.com"}, nil).Once()
				updatedInvite := *invite 
				updatedInvite.Used = true
				mockInviteRepo.On("Update", ctx, &updatedInvite).Return(nil, errors.New("update error")).Once()
			},
			expectedEmail: "",
			expectedErr:   config.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			email, err := uc.ValidateInvite(ctx, tt.token)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Equal(t, tt.expectedEmail, email)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEmail, email)
			}
		})
	}
}
