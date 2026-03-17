package psqlrepo

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type UserSettingsRepository interface {
	GetUserSettings(ctx context.Context, userID string) (*domain.UserSettings, error)
	CreateDefaultUserSettings(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error
	DeleteUserSettings(ctx context.Context, tx *sql.Tx, userID string) error
	UpdateUserSettings(ctx context.Context, settings *domain.UserSettings) (*domain.UserSettings, error)
}

type userSettingsRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewUserSettingsRepo(db *sqlx.DB, logger *zap.Logger) UserSettingsRepository {
	return &userSettingsRepository{
		db:     db,
		logger: logger,
	}
}

func (s *userSettingsRepository) DeleteUserSettings(ctx context.Context, tx *sql.Tx, userID string) error {
	const query = `DELETE FROM user_settings WHERE user_id = $1`
	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		s.logger.Error("failed to delete user settings", zap.String("user_id", userID), zap.Error(err))
		return err
	}

	return nil
}

func (s *userSettingsRepository) CreateDefaultUserSettings(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	log.Printf("Creating default user settings for user_id: %s", userID.String())

	const query = `INSERT INTO user_settings (user_id, notifications_enabled, language, timezone, privacy_level) VALUES ($1, $2, $3, $4, $5)`

	_, err := tx.ExecContext(ctx, query, userID, false, "en", "UTC", domain.PrivacyLevelPublic)
	if err != nil {
		s.logger.Error("failed to create default user settings", zap.String("user_id", userID.String()), zap.Error(err))
		return err
	}

	return nil
}

func (s *userSettingsRepository) GetUserSettings(ctx context.Context, userID string) (*domain.UserSettings, error) {
	const query = "select settings_id, user_id, notifications_enabled, language, timezone, privacy_level, updated_at from user_settings where user_id = $1"
	var settings domain.UserSettings
	err := s.db.GetContext(ctx, &settings, query, userID)
	log.Println()

	if err != nil {
		s.logger.Error("failed to get user settings", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	return &settings, nil
}

func (r *userSettingsRepository) UpdateUserSettings(
	ctx context.Context,
	settings *domain.UserSettings,
) (*domain.UserSettings, error) {

	query := "UPDATE user_settings SET "
	args := []interface{}{}
	argPos := 1

	if settings.Language != "" {
		query += fmt.Sprintf("language=$%d,", argPos)
		args = append(args, settings.Language)
		argPos++
	}

	if settings.Timezone != "" {
		query += fmt.Sprintf("timezone=$%d,", argPos)
		args = append(args, settings.Timezone)
		argPos++
	}

	if settings.PrivacyLevel != "" {
		query += fmt.Sprintf("privacy_level=$%d,", argPos)
		args = append(args, settings.PrivacyLevel)
		argPos++
	}

	query += fmt.Sprintf("notifications_enabled=$%d,", argPos)
	args = append(args, settings.NotificationsEnabled)
	argPos++

	query += "updated_at = NOW() "

	query += fmt.Sprintf("WHERE user_id=$%d ", argPos)
	args = append(args, settings.UserID)
	argPos++

	query += "RETURNING settings_id, user_id, notifications_enabled, language, timezone, privacy_level, updated_at"

	var updated domain.UserSettings

	err := r.db.GetContext(ctx, &updated, query, args...)
	if err != nil {
		r.logger.Error(
			"failed to update user settings",
			zap.String("user_id", settings.UserID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return &updated, nil
}
