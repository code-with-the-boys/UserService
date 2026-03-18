package server

import (
	"fmt"
	"time"

	service "github.com/code-with-the-boys/UserService/internal/services"
	"github.com/google/uuid"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func userSettingsToProto(dto *service.UserServiceUserSettings) *userServicepb.UserSettings {
	if dto == nil {
		return nil
	}

	var privacyLevel userServicepb.PrivacyLevel
	switch dto.PrivacyLevel {
	case "PUBLIC":
		privacyLevel = userServicepb.PrivacyLevel_PRIVACY_LEVEL_PUBLIC
	case "PRIVATE":
		privacyLevel = userServicepb.PrivacyLevel_PRIVACY_LEVEL_PRIVATE
	case "FRIENDS":
		privacyLevel = userServicepb.PrivacyLevel_PRIVACY_LEVEL_FRIENDS
	case "ONLY_ME":
		privacyLevel = userServicepb.PrivacyLevel_PRIVACY_LEVEL_ONLY_ME
	default:
		privacyLevel = userServicepb.PrivacyLevel_PRIVACY_LEVEL_PUBLIC
	}

	return &userServicepb.UserSettings{
		SettingsId:           dto.SettingsID,
		UserId:               dto.UserID,
		NotificationsEnabled: dto.NotificationsEnabled,
		Language:             dto.Language,
		Timezone:             dto.Timezone,
		PrivacyLevel:         privacyLevel,
		UpdatedAt:            timestamppb.New(dto.UpdatedAt),
	}
}

func protoToUserServiceDTO(pb *userServicepb.UpdateUserSettingsRequest) *service.UserServiceUserSettings {
	dto := &service.UserServiceUserSettings{}

	if pb.UserId != "" {
		dto.UserID = pb.UserId
	}
	if pb.NotificationsEnabled != nil {
		dto.NotificationsEnabled = *pb.NotificationsEnabled
	}

	if pb.Language != nil {
		dto.Language = *pb.Language
	}

	if pb.Timezone != nil {
		dto.Timezone = *pb.Timezone
	}

	if pb.PrivacyLevel != nil {
		switch *pb.PrivacyLevel {
		case userServicepb.PrivacyLevel_PRIVACY_LEVEL_PUBLIC:
			dto.PrivacyLevel = "PUBLIC"
		case userServicepb.PrivacyLevel_PRIVACY_LEVEL_PRIVATE:
			dto.PrivacyLevel = "PRIVATE"
		case userServicepb.PrivacyLevel_PRIVACY_LEVEL_FRIENDS:
			dto.PrivacyLevel = "FRIENDS"
		case userServicepb.PrivacyLevel_PRIVACY_LEVEL_ONLY_ME:
			dto.PrivacyLevel = "ONLY_ME"
		default:
			dto.PrivacyLevel = ""
		}
	}

	return dto
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

func ProtoToService(req *userServicepb.CreateUserProfileRequest, userID uuid.UUID) (*service.UserProfile, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Surname == "" {
		return nil, fmt.Errorf("surname is required")
	}

	profile := &service.UserProfile{
		UserID:  userID,
		Name:    req.Name,
		SurName: req.Surname,
	}

	if req.Patronymic != nil {
		profile.Patronymic = *req.Patronymic
	}

	if req.DateOfBirth != nil {
		dob := req.DateOfBirth.AsTime()

		if dob.After(time.Now()) {
			return nil, fmt.Errorf("date_of_birth cannot be in the future")
		}

		profile.DateOfBirth = dob
	}

	if req.Gender != nil {
		profile.Gender = genderEnumToString(*req.Gender)
	}

	if req.HeightCm != nil {
		if *req.HeightCm < 50 || *req.HeightCm > 300 {
			return nil, fmt.Errorf("height_cm must be between 50 and 300")
		}
		profile.HeightCm = int(*req.HeightCm)
	}

	if req.WeightKg != nil {
		if *req.WeightKg < 20 || *req.WeightKg > 500 {
			return nil, fmt.Errorf("weight_kg must be between 20 and 500")
		}
		profile.WeightKg = *req.WeightKg
	}

	if req.FitnessGoal != nil {
		profile.FitnessGoal = fitnessGoalEnumToString(*req.FitnessGoal)
	}

	if req.ExperienceLevel != nil {
		profile.ExperienceLevel = experienceLevelEnumToString(*req.ExperienceLevel)
	}

	if req.HealthLimitations != nil {
		profile.HealthLimitations = *req.HealthLimitations
	}

	return profile, nil
}

func stringToGenderEnum(gender string) userServicepb.Gender {
	switch gender {
	case "MALE":
		return userServicepb.Gender_GENDER_MALE
	case "FEMALE":
		return userServicepb.Gender_GENDER_FEMALE
	case "OTHER":
		return userServicepb.Gender_GENDER_OTHER
	default:
		return userServicepb.Gender_GENDER_UNSPECIFIED
	}
}

func ServiceToProto(user *service.UserProfile) *userServicepb.UserProfile {
	if user == nil {
		return nil
	}

	resp := &userServicepb.UserProfile{
		ProfileId: user.ProfileID.String(),
		UserId:    user.UserID.String(),
		Name:      user.Name,
		Surname:   user.SurName,
	}

	if user.Patronymic != "" {
		resp.Patronymic = &user.Patronymic
	}

	if !user.DateOfBirth.IsZero() {
		resp.DateOfBirth = timestamppb.New(user.DateOfBirth)
	}

	if user.Gender != "" {
		g := stringToGenderEnum(user.Gender)
		if g != userServicepb.Gender_GENDER_UNSPECIFIED {
			resp.Gender = &g
		}
	}

	if user.HeightCm != 0 {
		v := int32(user.HeightCm)
		resp.HeightCm = &v
	}

	if user.WeightKg != 0 {
		resp.WeightKg = &user.WeightKg
	}

	if user.FitnessGoal != "" {
		if val, ok := userServicepb.FitnessGoal_value[user.FitnessGoal]; ok {
			fg := userServicepb.FitnessGoal(val)
			resp.FitnessGoal = &fg
		}
	}

	if user.ExperienceLevel != "" {
		if val, ok := userServicepb.ExperienceLevel_value[user.ExperienceLevel]; ok {
			el := userServicepb.ExperienceLevel(val)
			resp.ExperienceLevel = &el
		}
	}

	if user.HealthLimitations != "" {
		resp.HealthLimitations = &user.HealthLimitations
	}

	if !user.CreatedAt.IsZero() {
		resp.CreatedAt = timestamppb.New(user.CreatedAt)
	}

	if !user.UpdatedAt.IsZero() {
		resp.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}

	// Для DateOfBirth используем уточнённое назначение: если не задано, поле остаётся пустым
	if !user.DateOfBirth.IsZero() {
		resp.DateOfBirth = timestamppb.New(user.DateOfBirth)
	}

	return resp
}
