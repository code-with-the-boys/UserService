package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"
	"github.com/code-with-the-boys/UserService/internal/repository/redisRepo"
	service "github.com/code-with-the-boys/UserService/internal/services"
	gw "github.com/code-with-the-boys/UserService/internal/transport/gateway"
	"github.com/code-with-the-boys/UserService/internal/transport/handlers/server"
	"github.com/code-with-the-boys/UserService/internal/transport/interceptors"
	"github.com/code-with-the-boys/UserService/internal/transport/trainclient"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	"github.com/code-with-the-boys/UserService/pkg/database/psql"
	"github.com/code-with-the-boys/UserService/pkg/database/redis"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	trainpb "github.com/mihnpro/UserServiceProtos/gen/go/train_service_api/v1"
	userServicepb "github.com/mihnpro/UserServiceProtos/gen/go/userServicepb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	httpAddr := getenv("HTTP_ADDR", ":8080")
	grpcAddr := getenv("GRPC_ADDR", ":9090")
	trainTarget := getenv("TRAIN_GRPC_ADDR", "127.0.0.1:50051")

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

	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("failed to listen grpc", zap.String("addr", grpcAddr), zap.Error(err))
	}

	skipMethods := map[string]bool{
		"/userService.UserService/SignUp":       true,
		"/userService.UserService/Login":        true,
		"/userService.UserService/RefreshToken": true,
	}

	authInterceptor := interceptors.NewInterceptorAuth(logger, skipMethods, jwtService)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LoggerInterceptor(logger),
			interceptors.PrometheusInterceptor(),
			authInterceptor.AuthRequired(),
		),
	)

	userServicepb.RegisterUserServiceServer(grpcSrv, grpcServer)
	reflection.Register(grpcSrv)

	ctx := context.Background()

	mux := runtime.NewServeMux()

	if err := userServicepb.RegisterUserServiceHandlerServer(ctx, mux, grpcServer); err != nil {
		logger.Fatal("register user gateway", zap.Error(err))
	}

	trainConn, err := grpc.NewClient(trainTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("dial train", zap.String("addr", trainTarget), zap.Error(err))
	}
	defer trainConn.Close()

	trainGatewayClient := trainclient.NewSecuringTrainClient(trainpb.NewTrainServiceClient(trainConn))
	if err := trainpb.RegisterTrainServiceHandlerClient(ctx, mux, trainGatewayClient); err != nil {
		logger.Fatal("register train gateway", zap.Error(err))
	}

	mainMux := http.NewServeMux()
	paymentURL := strings.TrimSpace(os.Getenv("PAYMENT_SERVICE_HTTP_URL"))
	if paymentURL != "" {
		paymentHandler, err := gw.NewPaymentHTTPProxy(logger, paymentURL)
		if err != nil {
			logger.Fatal("payment proxy", zap.String("url", paymentURL), zap.Error(err))
		}
		mainMux.Handle("/api/v1/payment/", http.StripPrefix("/api/v1/payment", paymentHandler))
		logger.Info("payment reverse proxy enabled", zap.String("target", paymentURL), zap.String("prefix", "/api/v1/payment"))
	}
	mainMux.Handle("/", mux)
	mainMux.Handle("/metrics", promhttp.Handler())

	httpHandler := gw.JWTHTTPMiddleware(logger, jwtService, mainMux)
	httpSrv := &http.Server{
		Addr:              httpAddr,
		Handler:           httpHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("gRPC listening", zap.String("addr", grpcLis.Addr().String()))
		if err := grpcSrv.Serve(grpcLis); err != nil {
			logger.Error("grpc serve", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("HTTP gateway listening", zap.String("addr", httpAddr))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http serve", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Info("shutdown signal")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	grpcSrv.GracefulStop()
	logger.Info("stopped")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
