
package context

import (
	"context"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authKey struct{}

type AuthData struct {
	User     *domain.User
	RawToken string
}

func WithAuth(ctx context.Context, auth AuthData) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

func GetAuth(ctx context.Context) (*AuthData, error) {
	auth, ok := ctx.Value(authKey{}).(AuthData)
	if !ok || auth.User == nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	return &auth, nil
}

func GetUser(ctx context.Context) (*domain.User, error) {
	auth, err := GetAuth(ctx)
	if err != nil {
		return nil, err
	}
	return auth.User, nil
}
