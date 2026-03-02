package model

import (
    "time"
    "github.com/google/uuid"
)

type UserSettings struct {
    SettingsID         uuid.UUID    `json:"settings_id" db:"settings_id"`
    UserID             uuid.UUID    `json:"user_id" db:"user_id"`
    NotificationsEnabled bool       `json:"notifications_enabled" db:"notifications_enabled"`
    Language           string       `json:"language" db:"language"`
    Timezone           string       `json:"timezone" db:"timezone"`
    PrivacyLevel       PrivacyLevel `json:"privacy_level" db:"privacy_level"`
    UpdatedAt          time.Time    `json:"updated_at" db:"updated_at"`
}

type PrivacyLevel string

const (
    PrivacyLevelPublic  PrivacyLevel = "PUBLIC"
    PrivacyLevelPrivate PrivacyLevel = "PRIVATE"
    PrivacyLevelFriends PrivacyLevel = "FRIENDS"
    PrivacyLevelOnlyMe  PrivacyLevel = "ONLY_ME"
)