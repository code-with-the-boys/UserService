package service

import (
	"context"
	"time"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/code-with-the-boys/UserService/internal/customErrors"
	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserServiceUserSettings struct {
	SettingsID           string    `json:"settings_id"`
	UserID               string    `json:"user_id"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
	Language             string    `json:"language"`
	Timezone             string    `json:"timezone"`
	PrivacyLevel         string    `json:"privacy_level"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type UserSettingsService interface {
	GetUserSettings(ctx context.Context, userID string) (*UserServiceUserSettings, error)
	UpdateUserSettings(ctx context.Context, userSett *UserServiceUserSettings) (*UserServiceUserSettings, error)
}

type userSettingsService struct {
	repo   psqlrepo.UserSettingsRepository
	logger *zap.Logger
}

func NewUserSettingsService(repo psqlrepo.UserSettingsRepository, logger *zap.Logger) UserSettingsService {
	return &userSettingsService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userSettingsService) GetUserSettings(ctx context.Context, userID string) (*UserServiceUserSettings, error) {

	if userID == "" {
		return nil, customErrors.NewInvalidArgumentError("user_id is required")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, customErrors.NewInvalidArgumentError("invalid user_id format")
	}

	settings, err := s.repo.GetUserSettings(ctx, uid.String())
	if err != nil {
		s.logger.Error("failed to get user settings",
			zap.String("user_id", userID),
			zap.Error(err),
		)

		return nil, customErrors.NewInternalError("failed to get user settings")
	}

	if settings == nil {
		return nil, customErrors.NewNotFoundError("user settings not found")
	}

	return toDTO(settings), nil
}

func (s *userSettingsService) UpdateUserSettings(ctx context.Context, dto *UserServiceUserSettings) (*UserServiceUserSettings, error) {

	if dto.UserID == "" {
		return nil, customErrors.NewInvalidArgumentError("user_id is required")
	}

	if _, err := uuid.Parse(dto.UserID); err != nil {
		return nil, customErrors.NewInvalidArgumentError("invalid user_id")
	}

	if dto.SettingsID != "" {
		if _, err := uuid.Parse(dto.SettingsID); err != nil {
			return nil, customErrors.NewInvalidArgumentError("invalid settings_id")
		}
	}

	if dto.Language != "" && len(dto.Language) > 10 {
		return nil, customErrors.NewValidationError("language is too long")
	}

	if dto.Timezone != "" && len(dto.Timezone) > 50 {
		return nil, customErrors.NewValidationError("timezone is too long")
	}

	if dto.PrivacyLevel != "" {
		switch domain.PrivacyLevel(dto.PrivacyLevel) {

		case domain.PrivacyLevelPublic,
			domain.PrivacyLevelPrivate,
			domain.PrivacyLevelFriends,
			domain.PrivacyLevelOnlyMe:

		default:
			return nil, customErrors.NewValidationError("invalid privacy_level")
		}
	}

	domainSettings, err := toDomain(dto)
	if err != nil {
		return nil, customErrors.NewInvalidArgumentError("invalid request data")
	}

	updated, err := s.repo.UpdateUserSettings(ctx, domainSettings)
	if err != nil {

		s.logger.Error("failed to update user settings",
			zap.String("user_id", dto.UserID),
			zap.Error(err),
		)

		return nil, customErrors.NewInternalError("failed to update user settings")
	}

	return toDTO(updated), nil
}

func toDomain(dto *UserServiceUserSettings) (*domain.UserSettings, error) {

	userID, err := uuid.Parse(dto.UserID)
	if err != nil {
		return nil, err
	}

	var settingsID uuid.UUID
	if dto.SettingsID != "" {
		settingsID, err = uuid.Parse(dto.SettingsID)
		if err != nil {
			return nil, err
		}
	}

	return &domain.UserSettings{
		SettingsID:           settingsID,
		UserID:               userID,
		NotificationsEnabled: dto.NotificationsEnabled,
		Language:             dto.Language,
		Timezone:             dto.Timezone,
		PrivacyLevel:         domain.PrivacyLevel(dto.PrivacyLevel),
		UpdatedAt:            dto.UpdatedAt,
	}, nil
}

func toDTO(settings *domain.UserSettings) *UserServiceUserSettings {
	return &UserServiceUserSettings{
		SettingsID:           settings.SettingsID.String(),
		UserID:               settings.UserID.String(),
		NotificationsEnabled: settings.NotificationsEnabled,
		Language:             settings.Language,
		Timezone:             settings.Timezone,
		PrivacyLevel:         string(settings.PrivacyLevel),
		UpdatedAt:            settings.UpdatedAt,
	}
}

