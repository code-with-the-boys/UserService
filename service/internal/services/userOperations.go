package service

import (
	"context"
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
}

type userOperationsService struct {
	userOperationsRepo psqlrepo.UserOperationsRepo
	logger             zap.Logger
	jwt                auth.JwtAuth
	redisRepo          redisRepo.RefreshTokenRepo
}

type UserServiceUserInfoResponse struct {
	UserID              string    `json:"user_id"`
	Email               string    `json:"email"`
	Phone               string    `json:"phone"`
	IsActive            bool      `json:"is_active"`
	SubscriptionStatus  string    `json:"subscription_status,omitempty"`
	SubscriptionExpires time.Time `json:"subscription_expires,omitempty"`
}

func NewUserOperationsService(logger *zap.Logger, jwt auth.JwtAuth, userOperationsRepo psqlrepo.UserOperationsRepo) UserOperationsService {
	return &userOperationsService{
		userOperationsRepo: userOperationsRepo,
		logger:             *logger,
		jwt:                jwt,
	}
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
		IsActive:            user.IsActive,
		SubscriptionStatus:  toProtoSubscriptionStatus(user.SubscriptionStatus),
		SubscriptionExpires: toProtoTimestamp(user.SubscriptionExpires),
		CreatedAt:           toProtoTimestamp(&user.CreatedAt),
		UpdatedAt:           toProtoTimestamp(&user.UpdatedAt),
	}, nil

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
