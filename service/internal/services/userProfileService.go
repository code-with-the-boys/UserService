package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/code-with-the-boys/UserService/internal/customErrors"
	"github.com/code-with-the-boys/UserService/internal/domain"
	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"
	"github.com/google/uuid"
	"github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
)

type UserProfile struct {
	ProfileID         uuid.UUID `json:"profile_id"`
	UserID            uuid.UUID `json:"user_id"`
	Name              string    `json:"name"`
	SurName           string    `json:"surname"`
	Patronymic        string    `json:"patronymic"`
	DateOfBirth       time.Time `json:"date_of_birth,omitempty"`
	Gender            string    `json:"gender,omitempty"`
	HeightCm          int       `json:"height_cm,omitempty"`
	WeightKg          float64   `json:"weight_kg,omitempty" `
	FitnessGoal       string    `json:"fitness_goal,omitempty"`
	ExperienceLevel   string    `json:"experience_level,omitempty"`
	HealthLimitations string    `json:"health_limitations,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UserProfileService interface {
	CreateUserProfile(ctx context.Context, userProfile *UserProfile) (*UserProfile, error)
	GetUserProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, req *userServicepb.UpdateUserProfileRequest) (*UserProfile, error)
	DeleteUserProfile(ctx context.Context, userID string) error
}

type userProfileService struct {
	userProfileRepo psqlrepo.UserProfileRepository
	logger          *zap.Logger
}

func genderEnumToString(g userServicepb.Gender) string {
	switch g {
	case userServicepb.Gender_GENDER_MALE:
		return "MALE"
	case userServicepb.Gender_GENDER_FEMALE:
		return "FEMALE"
	case userServicepb.Gender_GENDER_OTHER:
		return "OTHER"
	default:
		return ""
	}
}

func fitnessGoalEnumToString(g userServicepb.FitnessGoal) string {
	switch g {
	case userServicepb.FitnessGoal_FITNESS_GOAL_WEIGHT_LOSS:
		return "WEIGHT_LOSS"
	case userServicepb.FitnessGoal_FITNESS_GOAL_MUSCLE_GAIN:
		return "MUSCLE_GAIN"
	case userServicepb.FitnessGoal_FITNESS_GOAL_ENDURANCE:
		return "ENDURANCE"
	case userServicepb.FitnessGoal_FITNESS_GOAL_GENERAL_HEALTH:
		return "GENERAL_HEALTH"
	case userServicepb.FitnessGoal_FITNESS_GOAL_MAINTENANCE:
		return "MAINTENANCE"
	default:
		return ""
	}
}

func experienceLevelEnumToString(e userServicepb.ExperienceLevel) string {
	switch e {
	case userServicepb.ExperienceLevel_EXPERIENCE_LEVEL_BEGINNER:
		return "BEGINNER"
	case userServicepb.ExperienceLevel_EXPERIENCE_LEVEL_INTERMEDIATE:
		return "INTERMEDIATE"
	case userServicepb.ExperienceLevel_EXPERIENCE_LEVEL_ADVANCED:
		return "ADVANCED"
	case userServicepb.ExperienceLevel_EXPERIENCE_LEVEL_PROFESSIONAL:
		return "PROFESSIONAL"
	default:
		return ""
	}
}

func NewUserProfileService(userProfileRepo psqlrepo.UserProfileRepository, logger *zap.Logger) UserProfileService {
	return &userProfileService{
		userProfileRepo: userProfileRepo,
		logger:          logger,
	}
}

func (s *userProfileService) CreateUserProfile(ctx context.Context, userProfile *UserProfile) (*UserProfile, error) {
	if userProfile == nil {
		return nil, customErrors.NewInvalidArgumentError("user profile is required")
	}

	if userProfile.UserID == uuid.Nil {
		return nil, customErrors.NewInvalidArgumentError("uuid is nil")
	}

	if strings.TrimSpace(userProfile.Name) == "" {
		return nil, customErrors.NewInvalidArgumentError("name is required")
	}

	if strings.TrimSpace(userProfile.SurName) == "" {
		return nil, customErrors.NewInvalidArgumentError("surname is required")
	}

	if userProfile.Gender != "" {
		switch userProfile.Gender {
		case "MALE", "FEMALE", "OTHER":
		default:
			return nil, customErrors.NewInternalError("incorrect gender")
		}
	}

	if userProfile.FitnessGoal != "" {
		switch userProfile.FitnessGoal {
		case "WEIGHT_LOSS", "MUSCLE_GAIN", "ENDURANCE", "GENERAL_HEALTH", "MAINTENANCE":
		default:
			return nil, customErrors.NewInternalError("incorrect fitness_goal")
		}
	}

	if userProfile.ExperienceLevel != "" {
		switch userProfile.ExperienceLevel {
		case "BEGINNER", "INTERMEDIATE", "ADVANCED", "PROFESSIONAL":
		default:
			return nil, customErrors.NewInternalError("incorrect experience_level")
		}
	}

	if userProfile.HeightCm < 0 {
		return nil, customErrors.NewInvalidArgumentError("height_cm must be positive")
	}

	if userProfile.WeightKg < 0 {
		return nil, customErrors.NewInvalidArgumentError("weight_kg must be positive")
	}

	domainProfile := &domain.UserProfile{
		ProfileID:  uuid.New(),
		UserID:     userProfile.UserID,
		Name:       strings.TrimSpace(userProfile.Name),
		SurName:    strings.TrimSpace(userProfile.SurName),
		Patronymic: userProfile.Patronymic,
	}

	if !userProfile.DateOfBirth.IsZero() {
		domainProfile.DateOfBirth = &userProfile.DateOfBirth
	}

	if userProfile.Gender != "" {
		g := domain.Gender(userProfile.Gender)
		domainProfile.Gender = &g
	}

	if userProfile.HeightCm != 0 {
		domainProfile.HeightCm = &userProfile.HeightCm
	}

	if userProfile.WeightKg != 0 {
		domainProfile.WeightKg = &userProfile.WeightKg
	}

	if userProfile.FitnessGoal != "" {
		fg := domain.FitnessGoal(userProfile.FitnessGoal)
		domainProfile.FitnessGoal = &fg
	}

	if userProfile.ExperienceLevel != "" {
		el := domain.ExperienceLevel(userProfile.ExperienceLevel)
		domainProfile.ExperienceLevel = &el
	}

	if userProfile.HealthLimitations != "" {
		domainProfile.HealthLimitations = &userProfile.HealthLimitations
	}

	err := s.userProfileRepo.CreateUserProfile(ctx, domainProfile)

	if err != nil {
		s.logger.Error("failed to create user profile",
			zap.Error(err),
			zap.String("user_id", userProfile.UserID.String()),
		)
		return nil, customErrors.NewInternalError("failed to create user profile")
	}

	savedProfile, err := s.userProfileRepo.GetUserProfile(ctx, userProfile.UserID.String())
	if err != nil {
		s.logger.Error("failed to reload created profile",
			zap.Error(err),
			zap.String("user_id", userProfile.UserID.String()),
		)
		return nil, customErrors.NewInternalError("failed to load created user profile")
	}

	result := *userProfile
	result.ProfileID = savedProfile.ProfileID
	result.CreatedAt = savedProfile.CreatedAt
	result.UpdatedAt = savedProfile.UpdatedAt

	if savedProfile.DateOfBirth != nil {
		result.DateOfBirth = *savedProfile.DateOfBirth
	}
	if savedProfile.Gender != nil {
		result.Gender = string(*savedProfile.Gender)
	}
	if savedProfile.HeightCm != nil {
		result.HeightCm = *savedProfile.HeightCm
	}
	if savedProfile.WeightKg != nil {
		result.WeightKg = *savedProfile.WeightKg
	}
	if savedProfile.FitnessGoal != nil {
		result.FitnessGoal = string(*savedProfile.FitnessGoal)
	}
	if savedProfile.ExperienceLevel != nil {
		result.ExperienceLevel = string(*savedProfile.ExperienceLevel)
	}
	if savedProfile.HealthLimitations != nil {
		result.HealthLimitations = *savedProfile.HealthLimitations
	}

	s.logger.Info("user profile created",
		zap.String("user_id", userProfile.UserID.String()),
		zap.String("profile_id", result.ProfileID.String()),
	)

	return &result, nil
}

func (s *userProfileService) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	if userID == "" {
		return nil, customErrors.NewInvalidArgumentError("user id is required")
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, customErrors.NewInvalidArgumentError("incorrect uuid")
	}

	profile, err := s.userProfileRepo.GetUserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, psqlrepo.ErrNotFound) {
			return nil, customErrors.NewNotFoundError("user profile not found")
		}
		s.logger.Error("repo error", zap.Error(err))
		return nil, customErrors.NewInternalError("failed to get user profile")
	}

	usr := &UserProfile{
		ProfileID:  profile.ProfileID,
		UserID:     profile.UserID,
		Name:       profile.Name,
		SurName:    profile.SurName,
		Patronymic: profile.Patronymic,
		CreatedAt:  profile.CreatedAt,
		UpdatedAt:  profile.UpdatedAt,
	}

	if profile.DateOfBirth != nil {
		usr.DateOfBirth = *profile.DateOfBirth
	}

	if profile.Gender != nil {
		usr.Gender = string(*profile.Gender)
	}

	if profile.HeightCm != nil {
		usr.HeightCm = *profile.HeightCm
	}

	if profile.WeightKg != nil {
		usr.WeightKg = *profile.WeightKg
	}

	if profile.FitnessGoal != nil {
		usr.FitnessGoal = string(*profile.FitnessGoal)
	}

	if profile.ExperienceLevel != nil {
		usr.ExperienceLevel = string(*profile.ExperienceLevel)
	}

	if profile.HealthLimitations != nil {
		usr.HealthLimitations = *profile.HealthLimitations
	}

	return usr, nil
}

func (s *userProfileService) UpdateUserProfile(ctx context.Context, req *userServicepb.UpdateUserProfileRequest) (*UserProfile, error) {
	if req == nil {
		return nil, customErrors.NewInvalidArgumentError("request is required")
	}

	if req.UserId == "" {
		return nil, customErrors.NewInvalidArgumentError("user_id is required")
	}

	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, customErrors.NewInvalidArgumentError("invalid user_id")
	}

	existing, err := s.userProfileRepo.GetUserProfile(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, psqlrepo.ErrNotFound) {
			return nil, customErrors.NewNotFoundError("user profile not found")
		}
		return nil, customErrors.NewInternalError("failed to fetch existing user profile")
	}

	update := &domain.UserProfile{
		UserID: existing.UserID,
	}
	updatedFields := 0

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" && *req.Name != existing.Name {
		update.Name = strings.TrimSpace(*req.Name)
		updatedFields++
	}

	if req.Surname != nil && strings.TrimSpace(*req.Surname) != "" && *req.Surname != existing.SurName {
		update.SurName = strings.TrimSpace(*req.Surname)
		updatedFields++
	}

	if req.Patronymic != nil && *req.Patronymic != existing.Patronymic {
		update.Patronymic = *req.Patronymic
		updatedFields++
	}

	if req.DateOfBirth != nil {
		dob := req.DateOfBirth.AsTime()
		if dob.After(time.Now()) {
			return nil, customErrors.NewInvalidArgumentError("date_of_birth cannot be in the future")
		}
		update.DateOfBirth = &dob
		updatedFields++
	}

	if req.Gender != nil {
		if *req.Gender != userServicepb.Gender_GENDER_UNSPECIFIED {
			value := genderEnumToString(*req.Gender)
			if value == "" {
				return nil, customErrors.NewInvalidArgumentError("invalid gender")
			}
			dGender := domain.Gender(value)
			if existing.Gender == nil || string(*existing.Gender) != value {
				update.Gender = &dGender
				updatedFields++
			}
		}
	}

	if req.HeightCm != nil {
		if *req.HeightCm <= 0 {
			return nil, customErrors.NewInvalidArgumentError("height_cm must be positive")
		}
		if existing.HeightCm == nil || int(*req.HeightCm) != *existing.HeightCm {
			height := int(*req.HeightCm)
			update.HeightCm = &height
			updatedFields++
		}
	}

	if req.WeightKg != nil {
		if *req.WeightKg <= 0 {
			return nil, customErrors.NewInvalidArgumentError("weight_kg must be positive")
		}
		if existing.WeightKg == nil || *req.WeightKg != *existing.WeightKg {
			weight := *req.WeightKg
			update.WeightKg = &weight
			updatedFields++
		}
	}

	if req.FitnessGoal != nil {
		if *req.FitnessGoal != userServicepb.FitnessGoal_FITNESS_GOAL_UNSPECIFIED {
			value := fitnessGoalEnumToString(*req.FitnessGoal)
			if value == "" {
				return nil, customErrors.NewInvalidArgumentError("invalid fitness_goal")
			}
			dFitnessGoal := domain.FitnessGoal(value)
			switch dFitnessGoal {
			case domain.FitnessGoalWeightLoss, domain.FitnessGoalMuscleGain, domain.FitnessGoalEndurance, domain.FitnessGoalGeneralHealth, domain.FitnessGoalMaintenance:
			default:
				return nil, customErrors.NewInvalidArgumentError("invalid fitness_goal")
			}
			if existing.FitnessGoal == nil || string(*existing.FitnessGoal) != value {
				update.FitnessGoal = &dFitnessGoal
				updatedFields++
			}
		}
	}

	if req.ExperienceLevel != nil {
		if *req.ExperienceLevel != userServicepb.ExperienceLevel_EXPERIENCE_LEVEL_UNSPECIFIED {
			value := experienceLevelEnumToString(*req.ExperienceLevel)
			if value == "" {
				return nil, customErrors.NewInvalidArgumentError("invalid experience_level")
			}
			dExperienceLevel := domain.ExperienceLevel(value)
			switch dExperienceLevel {
			case domain.ExperienceLevelBeginner, domain.ExperienceLevelIntermediate, domain.ExperienceLevelAdvanced, domain.ExperienceLevelProfessional:
			default:
				return nil, customErrors.NewInvalidArgumentError("invalid experience_level")
			}
			if existing.ExperienceLevel == nil || string(*existing.ExperienceLevel) != value {
				update.ExperienceLevel = &dExperienceLevel
				updatedFields++
			}
		}
	}

	if req.HealthLimitations != nil && *req.HealthLimitations != "" && (existing.HealthLimitations == nil || *existing.HealthLimitations != *req.HealthLimitations) {
		hl := *req.HealthLimitations
		update.HealthLimitations = &hl
		updatedFields++
	}

	if updatedFields == 0 {
		return nil, customErrors.NewInvalidArgumentError("no fields to update")
	}

	if err := s.userProfileRepo.UpdateUserProfile(ctx, update); err != nil {
		if errors.Is(err, psqlrepo.ErrNotFound) {
			return nil, customErrors.NewNotFoundError("user profile not found")
		}
		if errors.Is(err, psqlrepo.ErrNoFieldsToUpdate) {
			return nil, customErrors.NewInvalidArgumentError("no fields to update")
		}
		s.logger.Error("failed to update profile", zap.Error(err))
		return nil, customErrors.NewInternalError("failed to update user profile")
	}

	_, err = s.userProfileRepo.GetUserProfile(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, psqlrepo.ErrNotFound) {
			return nil, customErrors.NewNotFoundError("user profile not found")
		}
		return nil, customErrors.NewInternalError("failed to get updated user profile")
	}

	return s.GetUserProfile(ctx, req.UserId)
}

func (s *userProfileService) DeleteUserProfile(ctx context.Context, userID string) error {
	if userID == "" {
		return customErrors.NewInvalidArgumentError("user_id is required")
	}

	if _, err := uuid.Parse(userID); err != nil {
		return customErrors.NewInvalidArgumentError("invalid user_id")
	}

	err := s.userProfileRepo.DeleteUserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, psqlrepo.ErrNotFound) {
			return customErrors.NewNotFoundError("user profile not found")
		}
		s.logger.Error("failed to delete user profile", zap.String("user_id", userID), zap.Error(err))
		return customErrors.NewInternalError("failed to delete user profile")
	}

	return nil
}
