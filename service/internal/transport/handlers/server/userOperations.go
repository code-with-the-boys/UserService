package server

import (
	"context"
	"fmt"
	"time"

	customContext "github.com/code-with-the-boys/UserService/internal/context"
	service "github.com/code-with-the-boys/UserService/internal/services"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *UserServiceServer) GetUser(ctx context.Context, req *userServicepb.GetUserRequest) (*userServicepb.GetUserResponse, error) {

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

	user, err := s.userOperationsService.GetUserInfo(ctx, req.UserId)

	if err != nil {
		return nil, err
	}

	s.logger.Info("User info got succsefuly from db", zap.String("email", user.Email), zap.String("phone", *user.Phone))

	return user, nil
}

func (s *UserServiceServer) UpdateUser(ctx context.Context, req *userServicepb.UpdateUserRequest) (*userServicepb.UpdateUserResponse, error) {

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

	userInfo := &service.UserServiceUserInfo{
		UserID: req.UserId,
	}

	if req.Email != nil {
		userInfo.Email = *req.Email
	}

	if req.Phone != nil {
		userInfo.Phone = *req.Phone
	}

	userInfo.IsActive = req.IsActive

	if req.SubscriptionStatus != nil {
		userInfo.SubscriptionStatus = req.SubscriptionStatus.String()
	}

	if req.SubscriptionExpires != nil {
		userInfo.SubscriptionExpires = req.SubscriptionExpires.AsTime()
	}

	user, err := s.userOperationsService.UpdateUserInfo(ctx, userInfo)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserServiceServer) DeleteUser(ctx context.Context, req *userServicepb.DeleteUserRequest) (*emptypb.Empty, error) {
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

	err = s.userOperationsService.DeleteUserInfo(ctx, req.UserId)

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *UserServiceServer) GetUserProfile(ctx context.Context, req *userServicepb.GetUserProfileRequest) (*userServicepb.UserProfile, error) {
	return &userServicepb.UserProfile{}, nil
}

func (s *UserServiceServer) UpdateUserProfile(ctx context.Context, req *userServicepb.UpdateUserProfileRequest) (*userServicepb.UserProfile, error) {
	return &userServicepb.UserProfile{}, nil
}

func (s *UserServiceServer) DeleteUserProfile(ctx context.Context, req *userServicepb.DeleteUserProfileRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

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