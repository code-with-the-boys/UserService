package service

import (
	"context"
	"errors"
	"time"

	"github.com/code-with-the-boys/UserService/internal/customErrors"
	"github.com/code-with-the-boys/UserService/internal/domain"
	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"
	"github.com/code-with-the-boys/UserService/internal/repository/redisRepo"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserOperationsService interface {
	GetUserInfo(ctx context.Context, userID string) (*userServicepb.GetUserResponse, error)
	UpdateUserInfo(ctx context.Context, req *UserServiceUserInfo) (*userServicepb.UpdateUserResponse, error)
	DeleteUserInfo(ctx context.Context, userID string) error
}

type userOperationsService struct {
	userOperationsRepo psqlrepo.UserOperationsRepo
	authUserRepo       psqlrepo.AuthUserRepo
	logger             zap.Logger
	jwt                auth.JwtAuth
	redisRepo          redisRepo.RefreshTokenRepo
	validator          validationsStuff
}

type UserServiceUserInfo struct {
	UserID              string    `json:"user_id"`
	Email               string    `json:"email"`
	Phone               string    `json:"phone"`
	IsActive            *bool     `json:"is_active"`
	SubscriptionStatus  string    `json:"subscription_status,omitempty"`
	SubscriptionExpires time.Time `json:"subscription_expires,omitempty"`
}

func NewUserOperationsService(logger *zap.Logger, jwt auth.JwtAuth, userOperationsRepo psqlrepo.UserOperationsRepo, redisRepo redisRepo.RefreshTokenRepo, authUserRepo psqlrepo.AuthUserRepo) UserOperationsService {
	return &userOperationsService{
		userOperationsRepo: userOperationsRepo,
		logger:             *logger,
		jwt:                jwt,
		validator:          NewValidationsStuff(logger, authUserRepo),
		redisRepo:          redisRepo,
		authUserRepo:       authUserRepo,
	}
}

func (u *userOperationsService) DeleteUserInfo(ctx context.Context, userID string) error {

	existingUser, err := u.userOperationsRepo.FindUserByID(ctx, userID)
	if err != nil {
		u.logger.Warn("user not found for deletion",
			zap.String("user_id", userID),
			zap.Error(err))
		return customErrors.NewNotFoundError("user not found")
	}

	if existingUser == nil {
		u.logger.Warn("user not found for deletion",
			zap.String("user_id", userID))
		return customErrors.NewNotFoundError("user not found")
	}

	err = u.userOperationsRepo.DeleteUserByID(ctx, userID)
	if err != nil {
		u.logger.Error("failed to delete user",
			zap.String("user_id", userID),
			zap.Error(err))
		return customErrors.NewInternalError("failed to delete user")
	}

	err = u.redisRepo.DeleteByUserID(ctx, userID)
	if err != nil {
		u.logger.Error("failed to delete refresh tokens for user",
			zap.String("user_id", userID),
			zap.Error(err))
		return customErrors.NewInternalError("failed to delete refresh tokens for user")
	}

	u.logger.Info("user deleted successfully",
		zap.String("user_id", userID))

	return nil

}

func (u *userOperationsService) UpdateUserInfo(ctx context.Context, req *UserServiceUserInfo) (*userServicepb.UpdateUserResponse, error) {

	existingUser, err := u.userOperationsRepo.FindUserByID(ctx, req.UserID)
	if err != nil {
		u.logger.Warn("user not found",
			zap.String("user_id", req.UserID),
			zap.Error(err))
		return nil, customErrors.NewNotFoundError("user not found")
	}

	updateData := &domain.User{
		UserID: existingUser.UserID,
	}

	hasChanges := false

	if req.Email != "" && req.Email != existingUser.Email {
		if err := u.validator.CheckEmailUniqueness(ctx, req.Email, req.UserID); err != nil {
			return nil, err
		}
		updateData.Email = req.Email
		hasChanges = true
	}

	if req.Phone != "" && req.Phone != *existingUser.Phone {
		if err := u.validator.CheckPhoneUniqueness(ctx, req.Phone, req.UserID); err != nil {
			return nil, err
		}
		updateData.Phone = &req.Phone
		hasChanges = true
	}

	if req.IsActive != existingUser.IsActive {
		updateData.IsActive = req.IsActive
		hasChanges = true
	}

	if req.SubscriptionStatus != "" && req.SubscriptionStatus != existingUser.SubscriptionStatus.String() {
		status, err := u.validator.parseSubscriptionStatus(req.SubscriptionStatus)
		if err != nil {
			return nil, customErrors.NewValidationError("invalid subscription status: " + req.SubscriptionStatus)
		}
		updateData.SubscriptionStatus = status
		hasChanges = true
	}

	if !req.SubscriptionExpires.IsZero() {
		if existingUser.SubscriptionExpires == nil || !req.SubscriptionExpires.Equal(*existingUser.SubscriptionExpires) {
			if req.SubscriptionExpires.Before(time.Now()) {
				return nil, customErrors.NewValidationError("subscription expiration date cannot be in the past")
			}
			updateData.SubscriptionExpires = &req.SubscriptionExpires
			hasChanges = true
		}
	} else if existingUser.SubscriptionExpires != nil {
		updateData.SubscriptionExpires = nil
		hasChanges = true
	}

	if !hasChanges {
		u.logger.Info("no fields to update",
			zap.String("user_id", req.UserID))
		return nil, customErrors.NewValidationError("no fields to update")
	}

	err = u.userOperationsRepo.UpdateUserInfo(ctx, updateData)
	if err != nil {
		u.logger.Error("failed to update user",
			zap.String("user_id", req.UserID),
			zap.Error(err))

		if errors.Is(err, psqlrepo.ErrDuplicateEmail) {
			return nil, customErrors.NewConflictError("email already exists")
		}
		if errors.Is(err, psqlrepo.ErrDuplicatePhone) {
			return nil, customErrors.NewConflictError("phone already exists")
		}
		return nil, customErrors.NewInternalError("failed to update user")
	}

	updatedUser, err := u.userOperationsRepo.FindUserByID(ctx, req.UserID)
	if err != nil {
		u.logger.Warn("failed to fetch updated user",
			zap.String("user_id", req.UserID),
			zap.Error(err))
		return nil, err
	}

	u.logger.Info("user updated successfully",
		zap.String("user_id", req.UserID),
		zap.Time("updated_at", updatedUser.UpdatedAt),
		zap.Any("changes", getChangedFields(updateData)))

	return &userServicepb.UpdateUserResponse{
		UserId:              updatedUser.UserID.String(),
		Email:               updatedUser.Email,
		Phone:               updatedUser.Phone,
		IsActive:            *updatedUser.IsActive,
		SubscriptionStatus:  toProtoSubscriptionStatus(updatedUser.SubscriptionStatus),
		SubscriptionExpires: toProtoTimestamp(updatedUser.SubscriptionExpires),
		UpdatedAt:           toProtoTimestamp(&updatedUser.UpdatedAt),
	}, nil
}
func (u *userOperationsService) GetUserInfo(ctx context.Context, userID string) (*userServicepb.GetUserResponse, error) {

	user, err := u.userOperationsRepo.FindUserByID(ctx, userID)
	if err != nil {
		u.logger.Warn("error while getting user info",
			zap.String("database_error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	return &userServicepb.GetUserResponse{
		UserId:              user.UserID.String(),
		Email:               user.Email,
		Phone:               user.Phone,
		IsActive:            *user.IsActive,
		SubscriptionStatus:  toProtoSubscriptionStatus(user.SubscriptionStatus),
		SubscriptionExpires: toProtoTimestamp(user.SubscriptionExpires),
		CreatedAt:           toProtoTimestamp(&user.CreatedAt),
		UpdatedAt:           toProtoTimestamp(&user.UpdatedAt),
	}, nil

}

func getChangedFields(updateData *domain.User) []string {
	var changes []string
	if updateData.Email != "" {
		changes = append(changes, "email")
	}
	if updateData.Phone != nil {
		changes = append(changes, "phone")
	}
	if updateData.IsActive != nil {
		changes = append(changes, "is_active")
	}
	if updateData.SubscriptionStatus != "" {
		changes = append(changes, "subscription_status")
	}
	if updateData.SubscriptionExpires != nil {
		changes = append(changes, "subscription_expires")
	}
	return changes
}

func toProtoSubscriptionStatus(status domain.SubscriptionStatus) userServicepb.SubscriptionStatus {
	switch status {
	case "UNSPECIFIED":
		return userServicepb.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED
	case "ACTIVE":
		return userServicepb.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE
	case "INACTIVE":
		return userServicepb.SubscriptionStatus_SUBSCRIPTION_STATUS_INACTIVE
	case "TRIAL":
		return userServicepb.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIAL
	case "EXPIRED":
		return userServicepb.SubscriptionStatus_SUBSCRIPTION_STATUS_EXPIRED
	default:
		return userServicepb.SubscriptionStatus_SUBSCRIPTION_STATUS_NONE
	}
}

func toProtoTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}
