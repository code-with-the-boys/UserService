package server

import (
	"context"
	"fmt"
	"time"

	customContext "github.com/code-with-the-boys/UserService/internal/context"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *UserServiceServer) GetUserSettings(ctx context.Context, req *userServicepb.GetUserSettingsRequest) (*userServicepb.UserSettings, error) {
	userFromContext, err := customContext.GetUser(ctx)
	if err != nil {
		s.logger.Info("user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if userFromContext.UserID.String() != req.UserId {
		s.logger.Info("User authenticated", zap.String("User ID from context", userFromContext.UserID.String()), zap.String("User ID from request", req.UserId))
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Info("User authenticated", zap.String("User ID from context", userFromContext.UserID.String()), zap.String("User ID from request", req.UserId))

	defer func() {
		trailer := metadata.Pairs("timestamp", time.Now().Format(time.RFC822))
		grpc.SetTrailer(ctx, trailer)
	}()

	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		if vals, ok := md["timestamp"]; ok && len(vals) > 0 {
			s.logger.Info("timestamp metadata:")
			for i, v := range vals {
				s.logger.Info("Metadata from client", zapcore.Field{
					Key:    fmt.Sprintf("timestamp[%d]", i),
					Type:   zapcore.StringType,
					String: v,
				})
			}
			s.logger.Info("timestamp metadata end")
		}
	} else {
		s.logger.Info("timestamp metadata not found")
	}

	userSet, err := s.userSettingsService.GetUserSettings(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return userSettingsToProto(userSet), nil
}

func (s *UserServiceServer) UpdateUserSettings(ctx context.Context, req *userServicepb.UpdateUserSettingsRequest) (*userServicepb.UserSettings, error) {
	userFromContext, err := customContext.GetUser(ctx)
	if err != nil {
		s.logger.Info("user not authenticated")
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if userFromContext.UserID.String() != req.UserId {
		s.logger.Info("User authenticated", zap.String("User ID from context", userFromContext.UserID.String()), zap.String("User ID from request", req.UserId))
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Info("User authenticated", zap.String("User ID from context", userFromContext.UserID.String()), zap.String("User ID from request", req.UserId))

	defer func() {
		trailer := metadata.Pairs("timestamp", time.Now().Format(time.RFC822))
		grpc.SetTrailer(ctx, trailer)
	}()

	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		if vals, ok := md["timestamp"]; ok && len(vals) > 0 {
			s.logger.Info("timestamp metadata:")
			for i, v := range vals {
				s.logger.Info("Metadata from client", zapcore.Field{
					Key:    fmt.Sprintf("timestamp[%d]", i),
					Type:   zapcore.StringType,
					String: v,
				})
			}
			s.logger.Info("timestamp metadata end")
		}
	} else {
		s.logger.Info("timestamp metadata not found")
	}

	dto := protoToUserServiceDTO(req)
	updated, err := s.userSettingsService.UpdateUserSettings(ctx, dto)
	if err != nil {
		return nil, err
	}

	return userSettingsToProto(updated), nil
}
