package server

import (
	"context"
	"fmt"
	"time"
	service "github.com/code-with-the-boys/UserService/internal/services"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *UserServiceServer) SignUp(ctx context.Context, req *userServicepb.SignUpRequest) (*userServicepb.SignUpResponse, error) {
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

	userResponse, err := s.authUserService.CreateUser(ctx, &service.UserServiceSignUpRequest{
		Email:    req.Email,
		Password: req.Password,
		Phone:    *req.Phone,
	})

	if err != nil {
		return nil, err
	}

	header := metadata.New(map[string]string{"location": "SCH", "timestamp": time.Now().Format(time.RFC822)})

	grpc.SetHeader(ctx, header)

	s.logger.Info("User created", zap.String("user_id", userResponse.UserID))

	return &userServicepb.SignUpResponse{Message: "User created",
		UserId: userResponse.UserID}, nil
}

func (s *UserServiceServer) Login(ctx context.Context, req *userServicepb.LoginRequest) (*userServicepb.LoginResponse, error) {
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

	userResponse, err := s.authUserService.CheckUserForLogin(ctx, &service.UserServiceLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return nil, err
	}

	header := metadata.New(map[string]string{"location": "SCH", "timestamp": time.Now().Format(time.RFC822)})

	grpc.SetHeader(ctx, header)

	s.logger.Info("User created", zap.String("user_id", userResponse.UserID))

	return &userServicepb.LoginResponse{
		AccessToken:  userResponse.AccessToken,
		RefreshToken: userResponse.RefreshToken,
		UserId:       userResponse.UserID,
		Email:        req.Email,
	}, nil
}

func (s *UserServiceServer) RefreshToken(ctx context.Context, req *userServicepb.RefreshTokenRequest) (*userServicepb.LoginResponse, error) {
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

	userResponse, err := s.authUserService.RefreshTokens(ctx, req.RefreshToken)

	if err != nil {
		return nil, err
	}

	header := metadata.New(map[string]string{"location": "SCH", "timestamp": time.Now().Format(time.RFC822)})

	grpc.SetHeader(ctx, header)

	s.logger.Info("User created", zap.String("user_id", userResponse.UserID))

	return &userServicepb.LoginResponse{
		UserId:       userResponse.UserID,
		Email:        userResponse.Email,
		AccessToken:  userResponse.AccessToken,
		RefreshToken: userResponse.RefreshToken,
	}, nil

}

func (s *UserServiceServer) Logout(ctx context.Context, req *userServicepb.LogoutRequest) (*emptypb.Empty, error) {

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

	err := s.authUserService.Logout(ctx, req.RefreshToken)

	if err != nil {
		return nil, err
	}

	s.logger.Info("User logged out successfully", zap.String("email", req.RefreshToken))

	return &emptypb.Empty{}, nil
}
