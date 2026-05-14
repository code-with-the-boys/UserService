package psqlrepo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Mock for UserSettingsRepository
type mockUserSettingsRepo struct{}

func (m *mockUserSettingsRepo) GetUserSettings(ctx context.Context, userID string) (*domain.UserSettings, error) {
	return nil, nil
}

func (m *mockUserSettingsRepo) CreateDefaultUserSettings(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	return nil
}

func (m *mockUserSettingsRepo) DeleteUserSettings(ctx context.Context, tx *sql.Tx, userID string) error {
	return nil
}

func (m *mockUserSettingsRepo) UpdateUserSettings(ctx context.Context, settings *domain.UserSettings) (*domain.UserSettings, error) {
	return nil, nil
}

func setupAuthUserRepo() (*authUserRepo, sqlmock.Sqlmock) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := zap.NewNop()
	mockSettings := &mockUserSettingsRepo{}
	repo := &authUserRepo{
		db:               sqlxDB,
		logger:           logger,
		userSettingsRepo: mockSettings,
	}
	return repo, mock
}

func TestAuthUserRepo_FindUserByEmail(t *testing.T) {
	repo, mock := setupAuthUserRepo()

	tests := []struct {
		name        string
		email       string
		mockSetup   func()
		expectedErr error
		expectedNil bool
	}{
		{
			name:  "user found",
			email: "test@example.com",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"user_id", "email", "phone", "password", "created_at", "updated_at", "is_active", "subscription_status", "subscription_expires"}).
					AddRow(uuid.New(), "test@example.com", "71234567890", "hashedpass", time.Now(), time.Now(), true, "ACTIVE", nil)
				mock.ExpectQuery(`SELECT \* FROM users WHERE email = \$1`).WithArgs("test@example.com").WillReturnRows(rows)
			},
			expectedErr: nil,
			expectedNil: false,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT \* FROM users WHERE email = \$1`).WithArgs("notfound@example.com").WillReturnError(sql.ErrNoRows)
			},
			expectedErr: ErrNotFound,
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := repo.FindUserByEmail(context.Background(), tt.email)
			if tt.expectedErr != nil {
				if err == nil || err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.expectedNil && user != nil {
				t.Errorf("expected nil user, got %v", user)
			}
			if !tt.expectedNil && user == nil {
				t.Errorf("expected user, got nil")
			}
		})
	}
}

func TestAuthUserRepo_FindUserByPhone(t *testing.T) {
	repo, mock := setupAuthUserRepo()

	tests := []struct {
		name        string
		phone       string
		mockSetup   func()
		expectedErr error
		expectedNil bool
	}{
		{
			name:  "user found",
			phone: "71234567890",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"user_id", "email", "phone", "password", "created_at", "updated_at", "is_active", "subscription_status", "subscription_expires"}).
					AddRow(uuid.New(), "test@example.com", "71234567890", "hashedpass", time.Now(), time.Now(), true, "ACTIVE", nil)
				mock.ExpectQuery(`SELECT \* FROM users WHERE phone = \$1`).WithArgs("71234567890").WillReturnRows(rows)
			},
			expectedErr: nil,
			expectedNil: false,
		},
		{
			name:  "user not found",
			phone: "79999999999",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT \* FROM users WHERE phone = \$1`).WithArgs("79999999999").WillReturnError(sql.ErrNoRows)
			},
			expectedErr: ErrNotFound,
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := repo.FindUserByPhone(context.Background(), tt.phone)
			if tt.expectedErr != nil {
				if err == nil || err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.expectedNil && user != nil {
				t.Errorf("expected nil user, got %v", user)
			}
			if !tt.expectedNil && user == nil {
				t.Errorf("expected user, got nil")
			}
		})
	}
}

func TestAuthUserRepo_CreateUser(t *testing.T) {
	repo, mock := setupAuthUserRepo()

	tests := []struct {
		name        string
		user        *domain.User
		mockSetup   func()
		expectedErr error
	}{
		{
			name: "successful creation",
			user: &domain.User{
				Email:    "new@example.com",
				Phone:    stringPtr("71234567890"),
				Password: "hashedpass",
			},
			mockSetup: func() {
				userID := uuid.New()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO users`).WithArgs("new@example.com", "71234567890", "hashedpass").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(userID))
				mock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name: "insert error",
			user: &domain.User{
				Email:    "new@example.com",
				Phone:    stringPtr("71234567890"),
				Password: "hashedpass",
			},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO users`).WithArgs("new@example.com", "71234567890", "hashedpass").WillReturnError(sqlmock.ErrCancelled)
				mock.ExpectRollback()
			},
			expectedErr: sqlmock.ErrCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			_, err := repo.CreateUser(context.Background(), tt.user)
			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
