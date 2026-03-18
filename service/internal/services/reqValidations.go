package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/code-with-the-boys/UserService/internal/domain"
	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"

	"github.com/code-with-the-boys/UserService/internal/customErrors"
	"go.uber.org/zap"
)

type validationsStuff struct {
	logger       *zap.Logger
	authUserRepo psqlrepo.AuthUserRepo
}

func NewValidationsStuff(logger *zap.Logger, authUserRepo psqlrepo.AuthUserRepo) validationsStuff {
	return validationsStuff{
		logger:       logger,
		authUserRepo: authUserRepo,
	}
}

func (s validationsStuff) ValidateEmailAndPhone(email, phone string) error {
	if email == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "email"),
			zap.String("validation_error", "email is empty"),
		)
		return customErrors.NewInvalidArgumentError("Empty email")
	}

	if phone == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "phone"),
			zap.String("validation_error", "phone is empty"),
		)
		return customErrors.NewInvalidArgumentError("Empty phone")
	}

	if err := s.ValidatePhone(phone); err != nil {
		return err
	}

	if err := s.ValidateEmail(email); err != nil {
		return err
	}

	return nil
}

func (s validationsStuff) ValidateEmail(email string) error {
	if email == "" {
		s.logger.Warn("validation error",
			zap.String("field", "email"),
			zap.String("validation_error", "email is empty"),
		)
		return customErrors.NewInvalidArgumentError("Email cannot be empty")
	}

	if !isValidEmail(email) {
		s.logger.Warn("validation error",
			zap.String("field", "email"),
			zap.String("email", email),
			zap.String("validation_error", "email format is invalid"),
		)
		return customErrors.NewInvalidArgumentError("Invalid email format")
	}

	return nil
}

func (s validationsStuff) ValidatePhone(phone string) error {
	if phone == "" {
		s.logger.Warn("validation error",
			zap.String("field", "phone"),
			zap.String("validation_error", "phone is empty"),
		)
		return customErrors.NewInvalidArgumentError("Phone cannot be empty")
	}

	normalizedPhone := normalizePhone(phone)

	if !isAllDigits(normalizedPhone) {
		s.logger.Warn("validation error",
			zap.String("field", "phone"),
			zap.String("phone", phone),
			zap.String("validation_error", "phone contains invalid characters"),
		)
		return customErrors.NewInvalidArgumentError("Phone must contain only digits")
	}

	if len(normalizedPhone) != 11 {
		s.logger.Warn("validation error",
			zap.String("field", "phone"),
			zap.String("phone", phone),
			zap.Int("length", len(normalizedPhone)),
			zap.String("validation_error", "phone must be 11 digits"),
		)
		return customErrors.NewInvalidArgumentError("Phone must be 11 digits long")
	}

	if normalizedPhone[0] != '7' && normalizedPhone[0] != '8' {
		s.logger.Warn("validation error",
			zap.String("field", "phone"),
			zap.String("phone", phone),
			zap.String("validation_error", "phone must start with 7 or 8"),
		)
		return customErrors.NewInvalidArgumentError("Phone must start with 7 or 8")
	}

	return nil
}

func (s validationsStuff) ValidateSubscriptionStatus(status string) error {
	validStatuses := map[string]bool{
		"UNSPECIFIED": true,
		"ACTIVE":      true,
		"INACTIVE":    true,
		"TRIAL":       true,
		"EXPIRED":     true,
		"NONE":        true,
	}

	if status == "" {
		s.logger.Warn("validation error",
			zap.String("field", "subscription_status"),
			zap.String("validation_error", "status is empty"),
		)
		return customErrors.NewInvalidArgumentError("Subscription status cannot be empty")
	}

	if !validStatuses[status] {
		s.logger.Warn("validation error",
			zap.String("field", "subscription_status"),
			zap.String("status", status),
			zap.String("validation_error", "invalid subscription status"),
		)
		return customErrors.NewInvalidArgumentError("Invalid subscription status")
	}

	return nil
}

func (s validationsStuff) ValidateSubscriptionExpires(expires time.Time) error {
	if expires.IsZero() {
		s.logger.Warn("validation error",
			zap.String("field", "subscription_expires"),
			zap.String("validation_error", "expiration date is zero"),
		)
		return customErrors.NewInvalidArgumentError("Subscription expiration date cannot be zero")
	}

	if expires.Before(time.Now()) {
		s.logger.Warn("validation error",
			zap.String("field", "subscription_expires"),
			zap.Time("expires", expires),
			zap.String("validation_error", "expiration date is in the past"),
		)
		return customErrors.NewInvalidArgumentError("Subscription expiration date cannot be in the past")
	}

	maxValidDate := time.Now().AddDate(10, 0, 0)
	if expires.After(maxValidDate) {
		s.logger.Warn("validation error",
			zap.String("field", "subscription_expires"),
			zap.Time("expires", expires),
			zap.String("validation_error", "expiration date is too far in the future"),
		)
		return customErrors.NewInvalidArgumentError("Subscription expiration date cannot be more than 10 years in the future")
	}

	return nil
}

func (s validationsStuff) ValidateUserID(userID string) error {
	if userID == "" {
		s.logger.Warn("validation error",
			zap.String("field", "user_id"),
			zap.String("validation_error", "user_id is empty"),
		)
		return customErrors.NewInvalidArgumentError("User ID cannot be empty")
	}

	if len(userID) != 36 {
		s.logger.Warn("validation error",
			zap.String("field", "user_id"),
			zap.String("user_id", userID),
			zap.String("validation_error", "invalid UUID format"),
		)
		return customErrors.NewInvalidArgumentError("Invalid user ID format")
	}

	return nil
}

func isValidEmail(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}

	if !strings.Contains(email, ".") {
		return false
	}

	if strings.Contains(email, " ") {
		return false
	}

	if len(email) > 255 {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func normalizePhone(phone string) string {
	var result strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (s validationsStuff) ValidateUpdateRequest(userID string, email, phone *string, subscriptionStatus *string, subscriptionExpires *time.Time) error {

	if err := s.ValidateUserID(userID); err != nil {
		return err
	}

	if email == nil && phone == nil && subscriptionStatus == nil && subscriptionExpires == nil {
		s.logger.Warn("validation error",
			zap.String("validation_error", "no fields to update"),
		)
		return customErrors.NewInvalidArgumentError("No fields to update provided")
	}

	if email != nil {
		if err := s.ValidateEmail(*email); err != nil {
			return err
		}
	}

	if phone != nil {
		if err := s.ValidatePhone(*phone); err != nil {
			return err
		}
	}

	if subscriptionStatus != nil {
		if err := s.ValidateSubscriptionStatus(*subscriptionStatus); err != nil {
			return err
		}
	}

	if subscriptionExpires != nil {
		if err := s.ValidateSubscriptionExpires(*subscriptionExpires); err != nil {
			return err
		}
	}

	return nil
}

func (s validationsStuff) parseSubscriptionStatus(status string) (domain.SubscriptionStatus, error) {
	trimmedStatus := strings.TrimPrefix(status, "SUBSCRIPTION_STATUS_")

	s.logger.Debug("subscription status", zap.String("status", trimmedStatus))

	switch trimmedStatus {
	case "UNSPECIFIED":
		return domain.SubscriptionStatusUnspecified, nil
	case "ACTIVE":
		return domain.SubscriptionStatusActive, nil
	case "INACTIVE":
		return domain.SubscriptionStatusInactive, nil
	case "TRIAL":
		return domain.SubscriptionStatusTrial, nil
	case "EXPIRED":
		return domain.SubscriptionStatusExpired, nil
	case "NONE":
		return domain.SubscriptionStatusNone, nil
	default:
		return domain.SubscriptionStatusNone, fmt.Errorf("unknown subscription status: %s", status)
	}
}

func (s validationsStuff) CheckEmailUniqueness(ctx context.Context, email, userID string) error {
	existingUser, err := s.authUserRepo.FindUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, psqlrepo.ErrNotFound) {
		return customErrors.NewInternalError("failed to check email uniqueness")
	}
	if existingUser != nil && existingUser.UserID.String() != userID {
		return customErrors.NewConflictError("email already exists")
	}
	return nil
}

func (s validationsStuff) CheckPhoneUniqueness(ctx context.Context, phone, userID string) error {
	existingUser, err := s.authUserRepo.FindUserByPhone(ctx, phone)
	if err != nil && !errors.Is(err, psqlrepo.ErrNotFound) {
		return customErrors.NewInternalError("failed to check phone uniqueness")
	}
	if existingUser != nil && existingUser.UserID.String() != userID {
		return customErrors.NewConflictError("phone already exists")
	}
	return nil
}
