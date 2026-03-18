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
	DeleteByUserID(ctx context.Context, userID string) error
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

func (r *refreshTokenRepo) DeleteByUserID(ctx context.Context, userID string) error {

	userSetKey := getUserTokensKey(userID)

	tokens, err := r.client.SMembers(ctx, userSetKey).Result()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	pipe := r.client.TxPipeline()

	for _, token := range tokens {
		pipe.Del(ctx, getTokenKey(token))
	}

	pipe.Del(ctx, userSetKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("failed to delete user tokens", zap.Error(err))
		return err
	}

	r.logger.Info("deleted all refresh tokens for user", zap.String("user_id", userID))

	return nil
}

func (r *refreshTokenRepo) Store(ctx context.Context, userID string, token string, expiresIn time.Duration) error {

	tokenKey := getTokenKey(token)
	userSetKey := getUserTokensKey(userID)

	pipe := r.client.TxPipeline()

	pipe.Set(ctx, tokenKey, userID, expiresIn)
	pipe.SAdd(ctx, userSetKey, token)
	pipe.Expire(ctx, userSetKey, expiresIn)

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("failed to store refresh token", zap.Error(err))
		return err
	}

	return nil
}

func (r *refreshTokenRepo) GetUserID(ctx context.Context, token string) (string, error) {

	key := getTokenKey(token)

	userID, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("token not found")
	}

	if err != nil {
		r.logger.Error("failed to get userID", zap.Error(err))
		return "", err
	}

	return userID, nil
}

func (r *refreshTokenRepo) Delete(ctx context.Context, token string) error {

	tokenKey := getTokenKey(token)

	userID, err := r.client.Get(ctx, tokenKey).Result()
	if err != nil {
		return err
	}

	userSetKey := getUserTokensKey(userID)

	pipe := r.client.TxPipeline()

	pipe.Del(ctx, tokenKey)
	pipe.SRem(ctx, userSetKey, token)

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("failed to delete token", zap.Error(err))
		return err
	}

	return nil
}

func getTokenKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

func getUserTokensKey(userID string) string {
	return fmt.Sprintf("user_refresh_tokens:%s", userID)
}