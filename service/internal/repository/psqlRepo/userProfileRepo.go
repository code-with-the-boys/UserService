package psqlrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type UserProfileRepository interface {
	GetUserProfile(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateUserProfile(ctx context.Context, profile *domain.UserProfile) error
	DeleteUserProfile(ctx context.Context, userID string) error
	DeleteUserProfileInTx(ctx context.Context, userID string, tx *sql.Tx) error
	CreateUserProfile(ctx context.Context, userProfile *domain.UserProfile) error
}

type userProfileRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewUserProfileRepository(db *sqlx.DB, logger *zap.Logger) UserProfileRepository {
	return &userProfileRepository{
		db:     db,
		logger: logger,
	}
}

func (u *userProfileRepository) CreateUserProfile(ctx context.Context, userProfile *domain.UserProfile) error {
	const query = `INSERT INTO user_profiles (user_id, name, surname, patronymic, date_of_birth, gender, height_cm, weight_kg, fitness_goal, experience_level, health_limitations) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := u.db.ExecContext(ctx, query, userProfile.UserID, userProfile.Name, userProfile.SurName, userProfile.Patronymic, userProfile.DateOfBirth, userProfile.Gender, userProfile.HeightCm, userProfile.WeightKg, userProfile.FitnessGoal, userProfile.ExperienceLevel, userProfile.HealthLimitations)
	if err != nil {
		u.logger.Error("failed to create user profile", zap.String("user_id", userProfile.UserID.String()), zap.Error(err))
		return err
	}
	u.logger.Info("user profile created", zap.String("user_id", userProfile.UserID.String()))
	return nil
}

func (u *userProfileRepository) GetUserProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	const query = `SELECT profile_id, user_id, name, surname, patronymic, date_of_birth, gender, height_cm, weight_kg, fitness_goal, experience_level, health_limitations, created_at, updated_at FROM user_profiles WHERE user_id = $1`
	var profile domain.UserProfile
	err := u.db.GetContext(ctx, &profile, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.logger.Info("user profile not found", zap.String("user_id", userID))
			return nil, ErrNotFound
		}
		u.logger.Error("failed to get user profile", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return &profile, nil
}

func (u *userProfileRepository) UpdateUserProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := "UPDATE user_profiles SET "
	args := []interface{}{}
	argPos := 1

	if profile.Name != "" {
		query += fmt.Sprintf("name=$%d,", argPos)
		args = append(args, profile.Name)
		argPos++
	}

	if profile.SurName != "" {
		query += fmt.Sprintf("surname=$%d,", argPos)
		args = append(args, profile.SurName)
		argPos++
	}

	if profile.Patronymic != "" {
		query += fmt.Sprintf("patronymic=$%d,", argPos)
		args = append(args, profile.Patronymic)
		argPos++
	}

	if profile.DateOfBirth != nil {
		query += fmt.Sprintf("date_of_birth=$%d,", argPos)
		args = append(args, profile.DateOfBirth)
		argPos++
	}

	if profile.Gender != nil {
		query += fmt.Sprintf("gender=$%d,", argPos)
		args = append(args, profile.Gender)
		argPos++
	}

	if profile.HeightCm != nil {
		query += fmt.Sprintf("height_cm=$%d,", argPos)
		args = append(args, profile.HeightCm)
		argPos++
	}

	if profile.WeightKg != nil {
		query += fmt.Sprintf("weight_kg=$%d,", argPos)
		args = append(args, profile.WeightKg)
		argPos++
	}

	if profile.FitnessGoal != nil {
		query += fmt.Sprintf("fitness_goal=$%d,", argPos)
		args = append(args, profile.FitnessGoal)
		argPos++
	}

	if profile.ExperienceLevel != nil {
		query += fmt.Sprintf("experience_level=$%d,", argPos)
		args = append(args, profile.ExperienceLevel)
		argPos++
	}

	if profile.HealthLimitations != nil {
		query += fmt.Sprintf("health_limitations=$%d,", argPos)
		args = append(args, profile.HealthLimitations)
		argPos++
	}

	if len(args) == 0 {
		return ErrNoFieldsToUpdate
	}

	query += "updated_at = NOW() "
	query += fmt.Sprintf("WHERE user_id=$%d", argPos)
	args = append(args, profile.UserID)

	res, err := u.db.ExecContext(ctx, query, args...)
	if err != nil {
		u.logger.Error("failed to update user profile", zap.String("user_id", profile.UserID.String()), zap.Error(err))
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		u.logger.Error("failed to get rows affected", zap.String("user_id", profile.UserID.String()), zap.Error(err))
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (u *userProfileRepository) DeleteUserProfile(ctx context.Context, userID string) error {
	const query = `DELETE FROM user_profiles WHERE user_id = $1`
	res, err := u.db.ExecContext(ctx, query, userID)
	if err != nil {
		u.logger.Error("failed to delete user profile", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		u.logger.Error("failed to get rows affected on delete", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (u *userProfileRepository) DeleteUserProfileInTx(ctx context.Context, userID string, tx *sql.Tx) error {
	const query = `DELETE FROM user_profiles WHERE user_id = $1`
	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		u.logger.Error("failed to delete user profile", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	return nil

}
