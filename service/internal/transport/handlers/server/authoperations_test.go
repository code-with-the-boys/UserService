package server

import (
	"context"
	"testing"

	"github.com/code-with-the-boys/UserService/internal/customErrors"
	service "github.com/code-with-the-boys/UserService/internal/services"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func stringPtr(s string) *string {
	return &s
}

// Mock for AuthUserService
type mockAuthUserService struct {
	createUser        func(ctx context.Context, request *service.UserServiceSignUpRequest) (*service.UserServiceSignUpResponse, error)
	checkUserForLogin func(ctx context.Context, request *service.UserServiceLoginRequest) (*service.UserServiceLoginResponse, error)
	refreshTokens     func(ctx context.Context, refreshToken string) (*service.UserServiceRefreshTokenResponse, error)
	logout            func(ctx context.Context, refreshToken string) error
}

func (m *mockAuthUserService) CreateUser(ctx context.Context, request *service.UserServiceSignUpRequest) (*service.UserServiceSignUpResponse, error) {
	return m.createUser(ctx, request)
}

func (m *mockAuthUserService) CheckUserForLogin(ctx context.Context, request *service.UserServiceLoginRequest) (*service.UserServiceLoginResponse, error) {
	return m.checkUserForLogin(ctx, request)
}

func (m *mockAuthUserService) RefreshTokens(ctx context.Context, refreshToken string) (*service.UserServiceRefreshTokenResponse, error) {
	return m.refreshTokens(ctx, refreshToken)
}

func (m *mockAuthUserService) Logout(ctx context.Context, refreshToken string) error {
	return m.logout(ctx, refreshToken)
}

func setupUserServiceServer() (*UserServiceServer, *mockAuthUserService) {
	logger := zap.NewNop()
	mockService := &mockAuthUserService{}
	server := &UserServiceServer{
		logger:          logger,
		authUserService: mockService,
	}
	return server, mockService
}

func TestUserServiceServer_SignUp(t *testing.T) {
	server, mockService := setupUserServiceServer()

	tests := []struct {
		name        string
		req         *userServicepb.SignUpRequest
		mockSetup   func()
		expectedErr bool
	}{
		{
			name: "successful signup",
			req: &userServicepb.SignUpRequest{
				Email:    "test@example.com",
				Password: "password123",
				Phone:    stringPtr("71234567890"),
			},
			mockSetup: func() {
				mockService.createUser = func(ctx context.Context, request *service.UserServiceSignUpRequest) (*service.UserServiceSignUpResponse, error) {
					return &service.UserServiceSignUpResponse{
						UserID:  "user_id",
						Message: "User created successfully",
					}, nil
				}
			},
			expectedErr: false,
		},
		{
			name: "service error",
			req: &userServicepb.SignUpRequest{
				Email:    "test@example.com",
				Password: "password123",
				Phone:    stringPtr("71234567890"),
			},
			mockSetup: func() {
				mockService.createUser = func(ctx context.Context, request *service.UserServiceSignUpRequest) (*service.UserServiceSignUpResponse, error) {
					return nil, customErrors.NewInternalError("service error")
				}
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"timestamp": "test"}))
			resp, err := server.SignUp(ctx, tt.req)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if resp == nil || resp.UserId == "" {
					t.Errorf("expected valid response, got %v", resp)
				}
			}
		})
	}
}

func TestUserServiceServer_Login(t *testing.T) {
	server, mockService := setupUserServiceServer()

	tests := []struct {
		name        string
		req         *userServicepb.LoginRequest
		mockSetup   func()
		expectedErr bool
	}{
		{
			name: "successful login",
			req: &userServicepb.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				mockService.checkUserForLogin = func(ctx context.Context, request *service.UserServiceLoginRequest) (*service.UserServiceLoginResponse, error) {
					return &service.UserServiceLoginResponse{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
						UserID:       "user_id",
						Message:      "User logged in successfully",
					}, nil
				}
			},
			expectedErr: false,
		},
		{
			name: "service error",
			req: &userServicepb.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				mockService.checkUserForLogin = func(ctx context.Context, request *service.UserServiceLoginRequest) (*service.UserServiceLoginResponse, error) {
					return nil, customErrors.NewInvalidArgumentError("invalid credentials")
				}
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"timestamp": "test"}))
			resp, err := server.Login(ctx, tt.req)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if resp == nil || resp.AccessToken == "" {
					t.Errorf("expected valid response, got %v", resp)
				}
			}
		})
	}
}

func TestUserServiceServer_RefreshToken(t *testing.T) {
	server, mockService := setupUserServiceServer()

	tests := []struct {
		name        string
		req         *userServicepb.RefreshTokenRequest
		mockSetup   func()
		expectedErr bool
	}{
		{
			name: "successful refresh",
			req: &userServicepb.RefreshTokenRequest{
				RefreshToken: "refresh_token",
			},
			mockSetup: func() {
				mockService.refreshTokens = func(ctx context.Context, refreshToken string) (*service.UserServiceRefreshTokenResponse, error) {
					return &service.UserServiceRefreshTokenResponse{
						UserID:       "user_id",
						Email:        "test@example.com",
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
					}, nil
				}
			},
			expectedErr: false,
		},
		{
			name: "service error",
			req: &userServicepb.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			mockSetup: func() {
				mockService.refreshTokens = func(ctx context.Context, refreshToken string) (*service.UserServiceRefreshTokenResponse, error) {
					return nil, customErrors.NewUnauthenticatedError("invalid token")
				}
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"timestamp": "test"}))
			resp, err := server.RefreshToken(ctx, tt.req)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if resp == nil || resp.AccessToken == "" {
					t.Errorf("expected valid response, got %v", resp)
				}
			}
		})
	}
}
