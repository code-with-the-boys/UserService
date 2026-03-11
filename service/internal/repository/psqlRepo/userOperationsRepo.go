package psqlrepo

import (
	"context"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type UserOperationsRepo interface {
	FindUserByID(ctx context.Context, id string) (*domain.User, error)
}

type userOperationsRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewUserOperationsRepo(db *sqlx.DB, logger *zap.Logger) UserOperationsRepo {
	return &userOperationsRepo{db: db,
		logger: logger}
}

func (a *userOperationsRepo) FindUserByID(ctx context.Context, id string) (*domain.User, error) {
	const query = `SELECT * FROM users WHERE user_id = $1`
	var user domain.User
	err := a.db.GetContext(ctx, &user, query, id)
	a.logger.Info("find user by id", zap.String("id", id))
	if err != nil {
		a.logger.Error("failed to find user by id", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return &user, nil
}
