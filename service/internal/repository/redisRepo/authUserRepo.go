package redisRepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RefreshTokenRepo interface {
	Store(ctx context.Context, userID string, token string, expiresIn time.Duration) error
	GetUserID(ctx context.Context, token string) (string, error)
	Delete(ctx context.Context, token string) error
}

type refreshTokenRepo struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRefreshTokenRepo(client *redis.Client, logger *zap.Logger) RefreshTokenRepo {
	return &refreshTokenRepo{
		client: client,
		logger: logger,
	}
}

func (r *refreshTokenRepo) Store(ctx context.Context, userID string, token string, expiresIn time.Duration) error {
	key := getKey(token)

	err := r.client.Set(ctx, key, userID, expiresIn).Err()

	if err != nil {
		r.logger.Error("failed to store refresh token", zap.Error(err))
		return err
	}

	return nil
}

func (r *refreshTokenRepo) GetUserID(ctx context.Context, token string) (string, error) {
	key := getKey(token)

	userID, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		r.logger.Error("token not found", zap.Error(err))
		return "", errors.New("token not found")
	}
	return userID, nil
}

func (r *refreshTokenRepo) Delete(ctx context.Context, token string) error {
	key := getKey(token)

	err := r.client.Del(ctx, key).Err()

	if err != nil {
		r.logger.Error("failed to delete refresh token", zap.Error(err))
		return err
	}

	return nil
}

func getKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}
