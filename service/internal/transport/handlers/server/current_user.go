package server

import (
	"context"

	customContext "github.com/code-with-the-boys/UserService/internal/context"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *UserServiceServer) GetCurrentUserID(ctx context.Context, _ *emptypb.Empty) (*userServicepb.CurrentUserIDResponse, error) {
	u, err := customContext.GetUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	return &userServicepb.CurrentUserIDResponse{UserId: u.UserID.String()}, nil
}
