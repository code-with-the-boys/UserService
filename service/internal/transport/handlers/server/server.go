package server

import (
	service "github.com/code-with-the-boys/UserService/internal/services"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
)

type UserServiceServer struct {
	userServicepb.UnimplementedUserServiceServer
	logger                *zap.Logger
	authUserService       service.AuthUserService
	userOperationsService service.UserOperationsService
}

func NewUserServiceServer(loggerZ *zap.Logger, authUserService service.AuthUserService, userOperationsService service.UserOperationsService) *UserServiceServer {
	return &UserServiceServer{logger: loggerZ, authUserService: authUserService, userOperationsService: userOperationsService}
}
