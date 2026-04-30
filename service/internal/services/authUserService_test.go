package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/code-with-the-boys/UserService/internal/customErrors"
	"github.com/code-with-the-boys/UserService/internal/domain"
	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Mock implementations

type mockAuthUserRepo struct {
	findUserByEmail func(ctx context.Context, email string) (*domain.User, error)
	findUserByPhone func(ctx context.Context, phone string) (*domain.User, error)
	createUser      func(ctx context.Context, user *domain.User) (uuid.UUID, error)
}

func (m *mockAuthUserRepo) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.findUserByEmail(ctx, email)
}

func (m *mockAuthUserRepo) FindUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return m.findUserByPhone(ctx, phone)
}

func (m *mockAuthUserRepo) CreateUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	return m.createUser(ctx, user)
}

type mockRefreshTokenRepo struct {
	store          func(ctx context.Context, refreshToken string, userID string, expiration time.Duration) error
	getUserID      func(ctx context.Context, refreshToken string) (string, error)
	delete         func(ctx context.Context, refreshToken string) error
	deleteByUserID func(ctx context.Context, userID string) error
}

func (m *mockRefreshTokenRepo) Store(ctx context.Context, refreshToken string, userID string, expiration time.Duration) error {
	return m.store(ctx, refreshToken, userID, expiration)
}

func (m *mockRefreshTokenRepo) GetUserID(ctx context.Context, refreshToken string) (string, error) {
	return m.getUserID(ctx, refreshToken)
}

func (m *mockRefreshTokenRepo) Delete(ctx context.Context, refreshToken string) error {
	return m.delete(ctx, refreshToken)
}

func (m *mockRefreshTokenRepo) DeleteByUserID(ctx context.Context, userID string) error {
	return m.deleteByUserID(ctx, userID)
}

type mockJwtAuth struct {
	generateTokens func(customer *domain.User) (string, string, error)
	validateToken  func(tokenString string) (*domain.User, error)
	refreshTokens  func(refreshToken string) (string, string, error)
}

func (m *mockJwtAuth) GenerateTokens(customer *domain.User) (string, string, error) {
	return m.generateTokens(customer)
}

func (m *mockJwtAuth) ValidateToken(tokenString string) (*domain.User, error) {
	return m.validateToken(tokenString)
}

func (m *mockJwtAuth) RefreshTokens(refreshToken string) (string, string, error) {
	return m.refreshTokens(refreshToken)
}

func setupAuthUserService() (*authUserService, *mockAuthUserRepo, *mockRefreshTokenRepo, *mockJwtAuth) {
	logger := zap.NewNop()
	mockRepo := &mockAuthUserRepo{}
	mockRedis := &mockRefreshTokenRepo{}
	mockJwt := &mockJwtAuth{}
	service := &authUserService{
		authUserRepo: mockRepo,
		logger:       *logger,
		jwt:          mockJwt,
		redisRepo:    mockRedis,
		validator:    validationsStuff{logger: logger, authUserRepo: mockRepo},
	}
	return service, mockRepo, mockRedis, mockJwt
}

func TestAuthUserService_CreateUser(t *testing.T) {
	service, mockRepo, _, _ := setupAuthUserService()

	tests := []struct {
		name           string
		request        *UserServiceSignUpRequest
		mockSetup      func()
		expectedError  error
		expectedUserID string
	}{
		{
			name: "successful user creation",
			request: &UserServiceSignUpRequest{
				Email:    "test@example.com",
				Password: "password123",
				Phone:    "71234567890",
			},
			mockSetup: func() {
				mockRepo.findUserByEmail = func(ctx context.Context, email string) (*domain.User, error) {
					return nil, psqlrepo.ErrNotFound
				}
				mockRepo.findUserByPhone = func(ctx context.Context, phone string) (*domain.User, error) {
					return nil, psqlrepo.ErrNotFound
				}
				mockRepo.createUser = func(ctx context.Context, user *domain.User) (uuid.UUID, error) {
					return uuid.New(), nil
				}
			},
			expectedError:  nil,
			expectedUserID: "", // will be checked dynamically
		},
		{
			name: "user already exists by email",
			request: &UserServiceSignUpRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Phone:    "71234567890",
			},
			mockSetup: func() {
				mockRepo.findUserByEmail = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{Email: email}, nil
				}
			},
			expectedError: customErrors.NewAlreadyExistsError("User with this email already exists"),
		},
		{
			name: "empty password",
			request: &UserServiceSignUpRequest{
				Email:    "test@example.com",
				Password: "",
				Phone:    "71234567890",
			},
			mockSetup:     func() {},
			expectedError: customErrors.NewInvalidArgumentError("Empty password"),
		},
		{
			name: "password too short",
			request: &UserServiceSignUpRequest{
				Email:    "test@example.com",
				Password: "123",
				Phone:    "71234567890",
			},
			mockSetup:     func() {},
			expectedError: customErrors.NewInvalidArgumentError("Invalid password length less than 8 characters"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := service.CreateUser(context.Background(), tt.request)
			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if response == nil || response.UserID == "" {
					t.Errorf("expected valid response, got %v", response)
				}
			}
		})
	}
}

func TestAuthUserService_CheckUserForLogin(t *testing.T) {
	service, mockRepo, mockRedis, mockJwt := setupAuthUserService()

	tests := []struct {
		name          string
		request       *UserServiceLoginRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful login",
			request: &UserServiceLoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockRepo.findUserByEmail = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{
						UserID:   uuid.New(),
						Email:    email,
						Password: string(hashedPassword),
					}, nil
				}
				mockJwt.generateTokens = func(customer *domain.User) (string, string, error) {
					return "access_token", "refresh_token", nil
				}
				mockRedis.store = func(ctx context.Context, refreshToken string, userID string, expiration time.Duration) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			request: &UserServiceLoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.findUserByEmail = func(ctx context.Context, email string) (*domain.User, error) {
					return nil, psqlrepo.ErrNotFound
				}
			},
			expectedError: customErrors.NewNotFoundError("User not found"),
		},
		{
			name: "invalid password",
			request: &UserServiceLoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mockRepo.findUserByEmail = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{
						UserID:   uuid.New(),
						Email:    email,
						Password: string(hashedPassword),
					}, nil
				}
			},
			expectedError: customErrors.NewInvalidArgumentError("Password is not valid"),
		},
		{
			name: "empty email",
			request: &UserServiceLoginRequest{
				Email:    "",
				Password: "password123",
			},
			mockSetup:     func() {},
			expectedError: customErrors.NewInvalidArgumentError("Empty email"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := service.CheckUserForLogin(context.Background(), tt.request)
			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if response == nil || response.AccessToken == "" {
					t.Errorf("expected valid response, got %v", response)
				}
			}
		})
	}
}

func TestAuthUserService_RefreshTokens(t *testing.T) {
	service, _, mockRedis, mockJwt := setupAuthUserService()

	tests := []struct {
		name          string
		refreshToken  string
		mockSetup     func()
		expectedError error
	}{
		{
			name:         "successful refresh",
			refreshToken: "valid_refresh_token",
			mockSetup: func() {
				const userIDStr = "00000000-0000-0000-0000-000000000001"

				mockRedis.getUserID = func(ctx context.Context, refreshToken string) (string, error) {
					return userIDStr, nil
				}
				mockJwt.validateToken = func(tokenString string) (*domain.User, error) {
					return &domain.User{
						UserID: uuid.MustParse(userIDStr),
						Email:  "test@example.com",
					}, nil
				}
				mockJwt.generateTokens = func(customer *domain.User) (string, string, error) {
					return "new_access_token", "new_refresh_token", nil
				}
				mockRedis.delete = func(ctx context.Context, refreshToken string) error {
					return nil
				}
				mockRedis.store = func(ctx context.Context, refreshToken string, userID string, expiration time.Duration) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid_token",
			mockSetup: func() {
				mockRedis.getUserID = func(ctx context.Context, refreshToken string) (string, error) {
					return "", errors.New("not found")
				}
			},
			expectedError: customErrors.NewUnauthenticatedError("Invalid refresh token"),
		},
		{
			name:          "empty refresh token",
			refreshToken:  "",
			mockSetup:     func() {},
			expectedError: customErrors.NewInvalidArgumentError("Empty refresh token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			response, err := service.RefreshTokens(context.Background(), tt.refreshToken)
			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if response == nil || response.AccessToken == "" {
					t.Errorf("expected valid response, got %v", response)
				}
			}
		})
	}
}

func TestAuthUserService_Logout(t *testing.T) {
	service, _, mockRedis, _ := setupAuthUserService()

	tests := []struct {
		name          string
		refreshToken  string
		mockSetup     func()
		expectedError error
	}{
		{
			name:         "successful logout",
			refreshToken: "refresh_token",
			mockSetup: func() {
				mockRedis.delete = func(ctx context.Context, refreshToken string) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:         "delete error",
			refreshToken: "refresh_token",
			mockSetup: func() {
				mockRedis.delete = func(ctx context.Context, refreshToken string) error {
					return errors.New("delete error")
				}
			},
			expectedError: customErrors.NewInternalError("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.Logout(context.Background(), tt.refreshToken)
			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
