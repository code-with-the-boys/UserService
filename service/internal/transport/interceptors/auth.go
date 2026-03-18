package interceptors

import (
	"context"
	"strings"

	customContext "github.com/code-with-the-boys/UserService/internal/context"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)


type InterceptorsAuth interface {
	AuthRequired() grpc.UnaryServerInterceptor
}


type interceptorAuth struct {
	logger      *zap.Logger
	jwt         auth.JwtAuth
	skipMethods map[string]bool
}


func NewInterceptorAuth(
	logger *zap.Logger,
	skipMethods map[string]bool,
	jwt auth.JwtAuth,
	) InterceptorsAuth {
		if logger == nil {
		logger = zap.NewNop()
	}
	return &interceptorAuth{
		logger:      logger,
		jwt:         jwt,
		skipMethods: skipMethods,
	}
}


func (i *interceptorAuth) AuthRequired() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {

		if i.skipMethods[info.FullMethod] {
			i.logger.Debug("skipping auth for public method", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			i.logger.Warn("missing metadata in request")
			return nil, status.Error(codes.Unauthenticated, "missing request metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		tokenRaw := strings.TrimSpace(authHeaders[0])

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(strings.ToLower(tokenRaw), "bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format: expected 'Bearer <token>'")
		}
		token := strings.TrimPrefix(tokenRaw, bearerPrefix)
		token = strings.TrimPrefix(token, "bearer ")

		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "empty token")
		}

		user, err := i.jwt.ValidateToken(token)
		if err != nil {
			i.logger.Error("token validation failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		newCtx := customContext.WithAuth(ctx, customContext.AuthData{
			User:     user,
			RawToken: token,
		})

		i.logger.Debug("auth successful",
			zap.String("method", info.FullMethod),
			zap.String("user_id", user.UserID.String()),
		)

		return handler(newCtx, req)
	}
}
