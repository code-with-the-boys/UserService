package psqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type AuthUserRepo interface {
	CreateUser(ctx context.Context, req *domain.User) (uuid.UUID, error)
	FindUserByEmail(ctx context.Context, email string) (*domain.User, error)
	FindUserByPhone(ctx context.Context, phone string) (*domain.User, error)
}

type authUserRepo struct {
	db               *sqlx.DB
	logger           *zap.Logger
	userSettingsRepo UserSettingsRepository
}

func NewAuthUserRepo(db *sqlx.DB, logger *zap.Logger, userSettingsRepository UserSettingsRepository) AuthUserRepo {
	return &authUserRepo{db: db,
		logger:           logger,
		userSettingsRepo: userSettingsRepository}
}

func (a *authUserRepo) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `SELECT * FROM users WHERE email = $1`
	var user domain.User
	err := a.db.GetContext(ctx, &user, query, email)

	a.logger.Info("find user by email", zap.String("email", email))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.logger.Info("user not found by email", zap.String("email", email))
			return nil, ErrNotFound
		}
		a.logger.Error("failed to find user by email",
			zap.String("email", email),
			zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

func (a *authUserRepo) FindUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	const query = `SELECT * FROM users WHERE phone = $1`
	var user domain.User
	err := a.db.GetContext(ctx, &user, query, phone)

	a.logger.Info("find user by phone", zap.String("phone", phone))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.logger.Info("user not found by phone", zap.String("phone", phone))
			return nil, ErrNotFound
		}
		a.logger.Error("failed to find user by phone",
			zap.String("phone", phone),
			zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

func (a *authUserRepo) CreateUser(ctx context.Context, req *domain.User) (uuid.UUID, error) {
	const query = `INSERT INTO users (email, phone, password) VALUES ($1, $2, $3) RETURNING user_id`

	var userID uuid.UUID

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		a.logger.Error("failed to begin transaction",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return uuid.Nil, err
	}

	defer tx.Rollback()

	err = tx.QueryRowContext(
		ctx,
		query,
		req.Email,
		req.Phone,
		req.Password,
	).Scan(&userID)

	if err != nil {
		a.logger.Error("failed to create user",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return uuid.Nil, err
	}

	if err := a.userSettingsRepo.CreateDefaultUserSettings(ctx, tx, userID); err != nil {
		tx.Rollback()
		a.logger.Error("failed to create default user settings during user creation", zap.String("email", req.Email), zap.String("phone", *req.Phone), zap.Error(err))
		return uuid.Nil, err
	}

	if err := tx.Commit(); err != nil {
		a.logger.Error("failed to commit transaction",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return uuid.Nil, err
	}

	a.logger.Info("user created",
		zap.String("email", req.Email),
	)

	return userID, nil
}
