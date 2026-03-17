package server

import (
	service "github.com/code-with-the-boys/UserService/internal/services"
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