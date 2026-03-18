package psqlrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type UserOperationsRepo interface {
	FindUserByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUserInfo(ctx context.Context, updateData *domain.User) error
	DeleteUserByID(ctx context.Context, id string) error

}

type userOperationsRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
	userSettingsRepository UserSettingsRepository
	userProfileRepository UserProfileRepository
}

func NewUserOperationsRepo(db *sqlx.DB, logger *zap.Logger, userSettingsRepository UserSettingsRepository, userrofileRepository UserProfileRepository) UserOperationsRepo {
	return &userOperationsRepo{db: db,
		logger: logger,
		userSettingsRepository: userSettingsRepository,
		userProfileRepository: userrofileRepository,
	}
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

func (a *userOperationsRepo) DeleteUserByID(ctx context.Context, id string) error {
	const query = `DELETE FROM users WHERE user_id = $1`

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		a.logger.Error("failed to begin transaction for deleting user", zap.String("id", id), zap.Error(err))
		return err
	}
	defer tx.Rollback()

	if err := a.userSettingsRepository.DeleteUserSettings(ctx, tx, id); err != nil {
		a.logger.Error("failed to delete user settings during user deletion", zap.String("id", id), zap.Error(err))
		return err
	}

	if err := a.userProfileRepository.DeleteUserProfileInTx(ctx, id, tx); err != nil {
		a.logger.Error("failed to delete user profile during user deletion", zap.String("id", id), zap.Error(err))
		return err
	}

	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		a.logger.Error("failed to delete user", zap.String("id", id), zap.Error(err))
		return err
	}

	if err := tx.Commit(); err != nil {
		a.logger.Error("failed to commit transaction for deleting user", zap.String("id", id), zap.Error(err))
		return err
	}

	a.logger.Info("user deleted successfully", zap.String("id", id))

	return nil
}


func (a *userOperationsRepo) UpdateUserInfo(ctx context.Context, updateData *domain.User) error {

	query := `UPDATE users SET`
	var setClauses []string
	var args []interface{}
	argCounter := 1

	if updateData.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argCounter))
		args = append(args, updateData.Email)
		argCounter++
	}

	if updateData.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argCounter))
		args = append(args, *updateData.Phone)
		argCounter++
	}

	if updateData.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argCounter))
		args = append(args, *updateData.IsActive)
		argCounter++
	}

	if updateData.SubscriptionStatus != "" {
		setClauses = append(setClauses, fmt.Sprintf("subscription_status = $%d", argCounter))
		args = append(args, updateData.SubscriptionStatus)
		argCounter++
	}

	if updateData.SubscriptionExpires != nil {
		setClauses = append(setClauses, fmt.Sprintf("subscription_expires = $%d", argCounter))
		args = append(args, *updateData.SubscriptionExpires)
		argCounter++
	}

	if len(setClauses) == 0 {
		return errors.New("no fields to update")
	}

	query += " " + strings.Join(setClauses, ", ")
	query += fmt.Sprintf(" WHERE user_id = $%d", argCounter)
	args = append(args, updateData.UserID)

	a.logger.Debug("executing update query",
		zap.String("query", query),
		zap.Any("args", args))

	result, err := a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	a.logger.Info("user updated successfully",
		zap.String("user_id", updateData.UserID.String()),
		zap.Int64("rows_affected", rowsAffected))

	return nil
}
