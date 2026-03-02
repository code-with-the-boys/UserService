// domain/model/user.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID              uuid.UUID          `json:"user_id" db:"user_id"`
	Email               string             `json:"email" db:"email"`
	Phone               *string            `json:"phone,omitempty" db:"phone"`
	Password            string             `json:"-" db:"password"`
	CreatedAt           time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at" db:"updated_at"`
	IsActive            bool               `json:"is_active" db:"is_active"`
	SubscriptionStatus  SubscriptionStatus `json:"subscription_status" db:"subscription_status"`
	SubscriptionExpires *time.Time         `json:"subscription_expires,omitempty" db:"subscription_expires"`
}

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "ACTIVE"
	SubscriptionStatusInactive SubscriptionStatus = "INACTIVE"
	SubscriptionStatusExpired  SubscriptionStatus = "EXPIRED"
	SubscriptionStatusTrial    SubscriptionStatus = "TRIAL"
	SubscriptionStatusNone     SubscriptionStatus = "NONE"
)
