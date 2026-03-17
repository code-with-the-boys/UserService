package psqlrepo

// import (
// 	"context"

// 	"github.com/jmoiron/sqlx"
// 	"go.uber.org/zap"
// )

// type UserProfileRepository interface {
// 	GetUserProfile(ctx context.Context, userID string) (*UserProfile, error)
// 	UpdateUserProfile(ctx context.Context, profile *UserProfile) error
// 	CreateDefaultUserProfile(ctx context.Context, tx *sqlx.Tx, userID string) error
// }

// type userProfileRepository struct {
// 	db     *sqlx.DB
// 	logger *zap.Logger
// }


// func NewUserProfileRepository(db *sqlx.DB, logger *zap.Logger) UserProfileRepository {
// 	return &userProfileRepository{
// 		db:     db,
// 		logger: logger,
// 	}
// }

// func 