package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/code-with-the-boys/UserService/internal/customErrors"
	"github.com/code-with-the-boys/UserService/internal/domain"
	psqlrepo "github.com/code-with-the-boys/UserService/internal/repository/psqlRepo"
	"github.com/code-with-the-boys/UserService/internal/repository/redisRepo"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)


type UserServiceLogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserServiceRefreshTokenResponse struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserServiceSignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type UserServiceSignUpResponse struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

type UserServiceLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserServiceLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Message      string `json:"message"`
}

type AuthUserService interface {
	CreateUser(ctx context.Context, request *UserServiceSignUpRequest) (*UserServiceSignUpResponse, error)
	CheckUserForLogin(ctx context.Context, request *UserServiceLoginRequest) (*UserServiceLoginResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*UserServiceRefreshTokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authUserService struct {
	authUserRepo psqlrepo.AuthUserRepo
	logger       zap.Logger
	jwt          auth.JwtAuth
	redisRepo    redisRepo.RefreshTokenRepo
}

func NewAuthUserService(loggerZ *zap.Logger, repo psqlrepo.AuthUserRepo, redisRepo redisRepo.RefreshTokenRepo, jwt auth.JwtAuth) AuthUserService {
	return &authUserService{
		authUserRepo: repo,
		logger:       *loggerZ,
		jwt:          jwt,
		redisRepo:    redisRepo,
	}

}

func (s *authUserService) CreateUser(ctx context.Context, request *UserServiceSignUpRequest) (*UserServiceSignUpResponse, error) {

	if request.Email == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "email"),
			zap.String("validation_error", "email is empty"),
		)
		return nil, customErrors.NewInvalidArgumentError("Empty email")
	}

	if request.Password == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "password"),
			zap.String("validation_error", "password is empty"),
		)
		return nil, customErrors.NewInvalidArgumentError("Empty password")
	}

	if request.Phone == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "phone"),
			zap.String("validation_error", "phone is empty"),
		)
		return nil, customErrors.NewInvalidArgumentError("Empty phone")
	}

	if len(request.Phone) != 11 {
		s.logger.Warn("missing required field",
			zap.String("field", "phone"),
			zap.String("validation_error", "phone is not valid"),
		)
		return nil, customErrors.NewInvalidArgumentError("Invalid phone")
	}

	if len(request.Password) < 8 {
		s.logger.Warn("missing required field",
			zap.String("field", "password"),
			zap.String("validation_error", "password is not valid"),
		)
		return nil, customErrors.NewInvalidArgumentError("Invalid password")
	}

	if !isValidEmail(request.Email) {
		s.logger.Warn("missing required field",
			zap.String("field", "email"),
			zap.String("validation_error", "email is not valid"),
		)
		return nil, customErrors.NewInvalidArgumentError("Invalid email")
	}

	userByEmail, err := s.authUserRepo.FindUserByEmail(ctx, request.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Info("user not found by email, can proceed with creation",
				zap.String("email", request.Email))
		} else {
			s.logger.Error("error while finding user by email",
				zap.String("email", request.Email),
				zap.Error(err))
			return nil, customErrors.NewInternalError("Database error")
		}

	} else if userByEmail != nil {
		s.logger.Warn("user already exists",
			zap.String("email", request.Email))
		return nil, customErrors.NewAlreadyExistsError("User with this email already exists")
	}

	userByPhone, err := s.authUserRepo.FindUserByPhone(ctx, request.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Info("user not found by phone, can proceed with creation",
				zap.String("phone", request.Phone))
		} else {
			s.logger.Error("error while finding user by phone",
				zap.String("phone", request.Phone),
				zap.Error(err))
			return nil, customErrors.NewInternalError("Database error")
		}
	} else if userByPhone != nil {
		s.logger.Warn("user already exists",
			zap.String("phone", request.Phone))
		return nil, customErrors.NewAlreadyExistsError("User with this phone already exists")
	}

	hashedPassword, err := s.hashPassword(request.Password)
	if err != nil {
		s.logger.Warn("error while hashing password",
			zap.String("field", "password"),
			zap.String("password_hash_error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	userID, err := s.authUserRepo.CreateUser(ctx, &domain.User{
		Email:    request.Email,
		Password: hashedPassword,
		Phone:    &request.Phone,
	})

	if err != nil {
		s.logger.Warn("error while creating user",
			zap.String("database_error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	return &UserServiceSignUpResponse{
		UserID:  userID.String(),
		Message: "User created successfully",
	}, nil

}

func (s *authUserService) CheckUserForLogin(ctx context.Context, request *UserServiceLoginRequest) (*UserServiceLoginResponse, error) {

	if request.Email == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "email"),
			zap.String("validation_error", "email is empty"),
		)
		return nil, customErrors.NewInvalidArgumentError("Empty email")
	}

	if request.Password == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "password"),
			zap.String("validation_error", "password is empty"),
		)
		return nil, customErrors.NewInvalidArgumentError("Empty password")
	}

	if len(request.Password) < 8 {
		s.logger.Warn("missing required field",
			zap.String("field", "password"),
			zap.String("validation_error", "password is not valid"),
		)
		return nil, customErrors.NewInvalidArgumentError("Invalid password")
	}

	if !isValidEmail(request.Email) {
		s.logger.Warn("missing required field",
			zap.String("field", "email"),
			zap.String("validation_error", "email is not valid"),
		)
		return nil, customErrors.NewInvalidArgumentError("Invalid email")
	}

	userByEmail, err := s.authUserRepo.FindUserByEmail(ctx, request.Email)

	if err != nil {
		s.logger.Warn("error while finding user by email",
			zap.String("field", "email"),
			zap.String("database_error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	if userByEmail == nil {
		s.logger.Warn("user has not been registred before",
			zap.String("field", "email"),
			zap.String("validation_error", "email is not valid"),
		)
		return nil, customErrors.NewNotFoundError("User not found")
	}

	if err := s.checkPassword(request.Password, userByEmail.Password); err != nil {
		s.logger.Warn("user has not been registred before",
			zap.String("field", "password"),
			zap.String("validation_error", "password is not valid"),
		)
		return nil, customErrors.NewInvalidArgumentError("Password is not valid")
	}

	accessToken, refreshToken, err := s.jwt.GenerateTokens(&domain.User{
		UserID: userByEmail.UserID,
		Email:  userByEmail.Email,
	})

	if err != nil {
		s.logger.Warn("error while generating tokens",
			zap.String("database_error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	if err := s.redisRepo.Store(ctx, userByEmail.UserID.String(), refreshToken, time.Hour*24); err != nil {
		s.logger.Warn("error while storing refresh token",
			zap.String("database_error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	return &UserServiceLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userByEmail.UserID.String(),
		Message:      "User logged in successfully",
	}, nil

}

func (s *authUserService) RefreshTokens(ctx context.Context, refreshToken string) (*UserServiceRefreshTokenResponse, error) {
	if refreshToken == "" {
		s.logger.Warn("missing required field",
			zap.String("field", "refresh_token"),
			zap.String("validation_error", "refresh token is empty"),
		)
		return nil, customErrors.NewInvalidArgumentError("Empty refresh token")
	}

	userID, err := s.redisRepo.GetUserID(ctx, refreshToken)
	if err != nil {
		s.logger.Warn("refresh token not found or already used",
			zap.String("error", err.Error()),
		)
		return nil, customErrors.NewUnauthenticatedError("Invalid refresh token")
	}

	user, err := s.jwt.ValidateToken(refreshToken)
	if err != nil {
		s.logger.Warn("error while validating token",
			zap.String("validation_error", err.Error()),
		)

		s.redisRepo.Delete(ctx, refreshToken)
		return nil, customErrors.NewUnauthenticatedError("Invalid refresh token")
	}

	if user.UserID.String() != userID {
		s.logger.Warn("user ID mismatch",
			zap.String("token_user_id", user.UserID.String()),
			zap.String("redis_user_id", userID),
		)
		return nil, customErrors.NewUnauthenticatedError("Invalid refresh token")
	}

	accessToken, newRefreshToken, err := s.jwt.GenerateTokens(&domain.User{
		UserID: user.UserID,
		Email:  user.Email,
	})
	if err != nil {
		s.logger.Warn("error while generating tokens",
			zap.String("error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	if err := s.redisRepo.Delete(ctx, refreshToken); err != nil {
		s.logger.Warn("error while deleting old refresh token",
			zap.String("error", err.Error()),
		)
	}

	if err := s.redisRepo.Store(ctx, newRefreshToken, user.UserID.String(), time.Hour*24); err != nil {
		s.logger.Warn("error while storing new refresh token",
			zap.String("error", err.Error()),
		)
		return nil, customErrors.NewInternalError(err.Error())
	}

	return &UserServiceRefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserID:       user.UserID.String(),
		Email:        user.Email,
	}, nil
}

func (u *authUserService) Logout(ctx context.Context, refreshToken string) error {
	if err := u.redisRepo.Delete(ctx, refreshToken); err != nil {
		u.logger.Warn("error while deleting refresh token",
			zap.String("database_error", err.Error()),
		)
		return customErrors.NewInternalError(err.Error())
	}
	u.logger.Info("user logged out successfully",
		zap.String("email", refreshToken),
	)

	return nil
}


func (u *authUserService) checkPassword(passwordFromReq string, passwordFromDB string) error {

	if err := bcrypt.CompareHashAndPassword([]byte(passwordFromDB), []byte(passwordFromReq)); err != nil {
		return err
	}
	return nil

}

func (u *authUserService) hashPassword(password string) (string, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err

}

func isValidEmail(email string) bool {

	if !strings.Contains(email, "@") {
		return false
	}

	if !strings.Contains(email, ".") {
		return false
	}

	if strings.Contains(email, " ") {
		return false
	}

	if len(email) > 255 {
		return false
	}

	return true
}
