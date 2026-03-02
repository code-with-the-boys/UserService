package model

import (
	"time"

	"github.com/google/uuid"
)

type UserProfile struct {
	ProfileID         uuid.UUID        `json:"profile_id" db:"profile_id"`
	UserID            uuid.UUID        `json:"user_id" db:"user_id"`
	Name              string           `json:"name" db:"name"`
	SurName           string           `json:"surname" db:"surname"`
	Patronymic        string           `json:"patronymic" db:"patronymic"`
	DateOfBirth       *time.Time       `json:"date_of_birth,omitempty" db:"date_of_birth"`
	Gender            *Gender          `json:"gender,omitempty" db:"gender"`
	HeightCm          *int             `json:"height_cm,omitempty" db:"height_cm"`
	WeightKg          *float64         `json:"weight_kg,omitempty" db:"weight_kg"`
	FitnessGoal       *FitnessGoal     `json:"fitness_goal,omitempty" db:"fitness_goal"`
	ExperienceLevel   *ExperienceLevel `json:"experience_level,omitempty" db:"experience_level"`
	HealthLimitations *string          `json:"health_limitations,omitempty" db:"health_limitations"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

type FitnessGoal string

const (
	FitnessGoalWeightLoss    FitnessGoal = "WEIGHT_LOSS"
	FitnessGoalMuscleGain    FitnessGoal = "MUSCLE_GAIN"
	FitnessGoalEndurance     FitnessGoal = "ENDURANCE"
	FitnessGoalGeneralHealth FitnessGoal = "GENERAL_HEALTH"
	FitnessGoalMaintenance   FitnessGoal = "MAINTENANCE"
)

type ExperienceLevel string

const (
	ExperienceLevelBeginner     ExperienceLevel = "BEGINNER"
	ExperienceLevelIntermediate ExperienceLevel = "INTERMEDIATE"
	ExperienceLevelAdvanced     ExperienceLevel = "ADVANCED"
	ExperienceLevelProfessional ExperienceLevel = "PROFESSIONAL"
)
