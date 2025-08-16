package usecases

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/mocks"
	"SalaryAdvance/internal/services"
	"SalaryAdvance/pkg/config"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateInviteToken(email string, userID uint) (string, error) {
	args := m.Called(email, userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(token string) (map[string]interface{}, error) {
	args := m.Called(token)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockJWTService := &MockJWTService{}
	uc := NewAuthUseCase(mockUserRepo, mockJWTService)

	tests := []struct {
		name            string
		email           string
		password        string
		mockSetup       func()
		expectedAccess  string
		expectedRefresh string
		expectedErr     error
	}{
		{
			name:     "Valid credentials",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				hashedPassword, _ := services.HashPassword("password123")
				user := &domain.User{ID: 1, Email: "test@example.com", Password: hashedPassword}
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil).Once()
				mockJWTService.On("GenerateAccessToken", user).Return("access_token", nil).Once()
				mockJWTService.On("GenerateRefreshToken", user).Return("refresh_token", nil).Once()
			},
			expectedAccess:  "access_token",
			expectedRefresh: "refresh_token",
			expectedErr:     nil,
		},
		{
			name:     "User not found",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(nil, errors.New("not found")).Once()
			},
			expectedAccess:  "",
			expectedRefresh: "",
			expectedErr:     config.ErrNotFound,
		},
		{
			name:     "Incorrect password",
			email:    "test@example.com",
			password: "wrongpassword",
			mockSetup: func() {
				hashedPassword, _ := services.HashPassword("password123")
				user := &domain.User{ID: 1, Email: "test@example.com", Password: hashedPassword}
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil).Once()
			},
			expectedAccess:  "",
			expectedRefresh: "",
			expectedErr:     config.ErrUnauthorized,
		},
		{
			name:     "Access token generation failure",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				hashedPassword, _ := services.HashPassword("password123")
				user := &domain.User{ID: 1, Email: "test@example.com", Password: hashedPassword}
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil).Once()
				mockJWTService.On("GenerateAccessToken", user).Return("", errors.New("token error")).Once()
			},
			expectedAccess:  "",
			expectedRefresh: "",
			expectedErr:     config.ErrInternalServer,
		},
		{
			name:     "Refresh token generation failure",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func() {
				hashedPassword, _ := services.HashPassword("password123")
				user := &domain.User{ID: 1, Email: "test@example.com", Password: hashedPassword}
				mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil).Once()
				mockJWTService.On("GenerateAccessToken", user).Return("access_token", nil).Once()
				mockJWTService.On("GenerateRefreshToken", user).Return("", errors.New("token error")).Once()
			},
			expectedAccess:  "",
			expectedRefresh: "",
			expectedErr:     config.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			accessToken, refreshToken, err := uc.Login(ctx, tt.email, tt.password)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Equal(t, tt.expectedAccess, accessToken)
				assert.Equal(t, tt.expectedRefresh, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAccess, accessToken)
				assert.Equal(t, tt.expectedRefresh, refreshToken)
			}
		})
	}
}

func TestAuthUseCase_RegisterFromInvite(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockJWTService := &MockJWTService{}
	uc := NewAuthUseCase(mockUserRepo, mockJWTService)

	tests := []struct {
		name        string
		token       string
		password    string
		mockSetup   func()
		expectedErr error
	}{
		{
			name:     "Valid invite token and password",
			token:    "valid_token",
			password: "password123",
			mockSetup: func() {
				claims := map[string]interface{}{"email": "test@example.com"}
				mockJWTService.On("ValidateToken", "valid_token").Return(claims, nil).Once()
				hashedPassword, _ := services.HashPassword("password123")
				user := &domain.User{Email: "test@example.com", Password: hashedPassword, Role: "uploader"}
				mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(user, nil).Once()
			},
			expectedErr: nil,
		},
		{
			name:     "Invalid invite token",
			token:    "invalid_token",
			password: "password123",
			mockSetup: func() {
				mockJWTService.On("ValidateToken", "invalid_token").Return(map[string]interface{}{}, errors.New("invalid token")).Once()
			},
			expectedErr: config.ErrUnauthorized,
		},
		{
			name:     "Invalid email in token claims",
			token:    "valid_token",
			password: "password123",
			mockSetup: func() {
				claims := map[string]interface{}{"email": 123} 
				mockJWTService.On("ValidateToken", "valid_token").Return(claims, nil).Once()
			},
			expectedErr: config.ErrBadRequest,
		},
		{
			name:     "User creation failure",
			token:    "valid_token",
			password: "password123",
			mockSetup: func() {
				claims := map[string]interface{}{"email": "test@example.com"}
				mockJWTService.On("ValidateToken", "valid_token").Return(claims, nil).Once()
				mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil, errors.New("creation error")).Once()
			},
			expectedErr: config.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := uc.RegisterFromInvite(ctx, tt.token, tt.password)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUseCase_RefreshToken(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockJWTService := &MockJWTService{}
	uc := NewAuthUseCase(mockUserRepo, mockJWTService)

	tests := []struct {
		name           string
		refreshToken   string
		mockSetup      func()
		expectedAccess string
		expectedErr    error
	}{
		{
			name:         "Valid refresh token",
			refreshToken: "valid_refresh_token",
			mockSetup: func() {
				claims := map[string]interface{}{"id": 1.0}
				mockJWTService.On("ValidateToken", "valid_refresh_token").Return(claims, nil).Once()
				user := &domain.User{ID: 1}
				mockJWTService.On("GenerateAccessToken", user).Return("new_access_token", nil).Once()
			},
			expectedAccess: "new_access_token",
			expectedErr:    nil,
		},
		{
			name:         "Invalid refresh token",
			refreshToken: "invalid_refresh_token",
			mockSetup: func() {
				mockJWTService.On("ValidateToken", "invalid_refresh_token").Return(map[string]interface{}{}, errors.New("invalid token")).Once()
			},
			expectedAccess: "",
			expectedErr:    config.ErrUnauthorized,
		},
		{
			name:         "Invalid user ID in token claims",
			refreshToken: "valid_refresh_token",
			mockSetup: func() {
				claims := map[string]interface{}{"id": "invalid"} 
				mockJWTService.On("ValidateToken", "valid_refresh_token").Return(claims, nil).Once()
			},
			expectedAccess: "",
			expectedErr:    config.ErrBadRequest,
		},
		{
			name:         "Access token generation failure",
			refreshToken: "valid_refresh_token",
			mockSetup: func() {
				claims := map[string]interface{}{"id": 1.0}
				mockJWTService.On("ValidateToken", "valid_refresh_token").Return(claims, nil).Once()
				user := &domain.User{ID: 1}
				mockJWTService.On("GenerateAccessToken", user).Return("", errors.New("token error")).Once()
			},
			expectedAccess: "",
			expectedErr:    config.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			accessToken, err := uc.RefreshToken(ctx, tt.refreshToken)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Equal(t, tt.expectedAccess, accessToken)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAccess, accessToken)
			}
		})
	}
}
