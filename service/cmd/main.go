package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"
	"github.com/code-with-the-boys/UserService/internal/repository/redisRepo"
	service "github.com/code-with-the-boys/UserService/internal/services"
	"github.com/code-with-the-boys/UserService/internal/transport/handlers/server"
	"github.com/code-with-the-boys/UserService/internal/transport/interceptors"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	"github.com/code-with-the-boys/UserService/pkg/database/psql"
	"github.com/code-with-the-boys/UserService/pkg/database/redis"
	"github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = "8080"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	psql.Init()

	db := psql.GetDB()

	redis.Init()

	redisClient := redis.GetRedisClient()

	defer db.Close()
	defer redisClient.Close()

	userSettingsRepo := psqlrepo.NewUserSettingsRepo(db, logger)

	authUserRepo := psqlrepo.NewAuthUserRepo(db, logger, userSettingsRepo)

	authUserRepoRedis := redisRepo.NewRefreshTokenRepo(redisClient, logger)
	jwtService := auth.NewJwtService()

	userProfileRepo := psqlrepo.NewUserProfileRepository(db, logger)

	authUserService := service.NewAuthUserService(logger, authUserRepo, authUserRepoRedis, jwtService)
	userOperationsRepo := psqlrepo.NewUserOperationsRepo(db, logger, userSettingsRepo, userProfileRepo)
	userOperationsService := service.NewUserOperationsService(logger, jwtService, userOperationsRepo, authUserRepoRedis, authUserRepo)

	userSettingsService := service.NewUserSettingsService(userSettingsRepo, logger)

	userProfileService := service.NewUserProfileService(userProfileRepo, logger)

	grpcServer := server.NewUserServiceServer(logger, authUserService, userOperationsService, userSettingsService, userProfileService)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("failed to listen: " + err.Error())
	}

	skipMethods := map[string]bool{
		"/userService.UserService/SignUp":       true,
		"/userService.UserService/Login":        true,
		"/userService.UserService/RefreshToken": true,
	}

	authInterceptor := interceptors.NewInterceptorAuth(logger, skipMethods, jwtService)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.AuthRequired()),
	)

	userServicepb.RegisterUserServiceServer(s, grpcServer)
	reflection.Register(s)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("Graceful shoutdown")
		s.GracefulStop()
	}()

	logger.Info("User server listening at " + lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		logger.Error("failed to serve: " + err.Error())
	}

}
