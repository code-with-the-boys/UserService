package psqlrepo

import (
	"context"

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
	db     *sqlx.DB
	logger *zap.Logger
}

func NewAuthUserRepo(db *sqlx.DB, logger *zap.Logger) AuthUserRepo {
	return &authUserRepo{db: db,
		logger: logger}
}



func (a *authUserRepo) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `SELECT * FROM users WHERE email = $1`
	var user domain.User
	err := a.db.GetContext(ctx, &user, query, email)
	a.logger.Info("find user by email", zap.String("email", email))
	if err != nil {
		a.logger.Error("failed to find user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func (a *authUserRepo) FindUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	const query = `SELECT * FROM users WHERE phone = $1`
	var user domain.User
	err := a.db.GetContext(ctx, &user, query, phone)
	a.logger.Info("find user by phone", zap.String("phone", phone))
	if err != nil {
		a.logger.Error("failed to find user by phone", zap.String("phone", phone), zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func (a *authUserRepo) CreateUser(ctx context.Context, req *domain.User) (uuid.UUID, error) {
	const query = `INSERT INTO users (email, phone, password) VALUES ($1, $2, $3) RETURNING user_id`
	var userID uuid.UUID
	err := a.db.GetContext(ctx, &userID, query, req.Email, req.Phone, req.Password)
	a.logger.Info("create user", zap.String("email", req.Email), zap.String("phone", *req.Phone))
	if err != nil {
		a.logger.Error("failed to create user", zap.String("email", req.Email), zap.String("phone", *req.Phone), zap.Error(err))
		return uuid.Nil, err
	}
	return userID, nil
}

