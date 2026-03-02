package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtService struct {
	accessSecretKey  string
	refreshSecretKey string
}

func NewJwtService() JwtAuth {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

	return &jwtService{
		accessSecretKey:  accessSecret,
		refreshSecretKey: refreshSecret,
	}
}

type claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

func (j *jwtService) GenerateTokens(user *domain.User) (string, string, error) {
	accessToken, err := generateToken(user, "access", time.Minute*5, j.accessSecretKey)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken(user, "refresh", time.Hour*24*7, j.refreshSecretKey)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func generateToken(user *domain.User, tokenType string, duration time.Duration, secretKey string) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := &claims{
		UserID:    user.UserID,
		Email:     user.Email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "Shop-system",
			Subject:   tokenType,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func (j *jwtService) ValidateToken(tokenString string) (*domain.User, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &claims{})

	if err != nil {
		return nil, errors.New("Invalid token format")
	}

	claimsFromToken, ok := token.Claims.(*claims)
	if !ok {
		return nil, errors.New("Invalid token claims")
	}

	var secretKey string
	switch claimsFromToken.TokenType {
	case "access":
		secretKey = j.accessSecretKey
	case "refresh":
		secretKey = j.refreshSecretKey
	default:
		return nil, errors.New("Invalid token type")
	}

	validToken, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid signing method")
		}

		return []byte(secretKey), nil
	})

	if err != nil || !validToken.Valid {
		return nil, errors.New("Invalid token")
	}

	validClaims, ok := validToken.Claims.(*claims)
	if !ok {
		return nil, errors.New("Invalid token claims")
	}

	return &domain.User{
		UserID: validClaims.UserID,
		Email:  validClaims.Email,
	}, nil

}

func (j *jwtService) RefreshTokens(refreshToken string) (string, string, error) {

	customer, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", "", errors.New("Cannot refresh invalid token")
	}

	log.Println("Token starts to refresh")

	token, _, _ := new(jwt.Parser).ParseUnverified(refreshToken, &claims{})
	if claimsFromToken, ok := token.Claims.(*claims); ok && claimsFromToken.TokenType != "refresh" {
		return "", "", errors.New("Cannot refresh access token")
	}

	return j.GenerateTokens(customer)
}
