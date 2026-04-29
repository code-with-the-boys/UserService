package auth

import (
	"github.com/code-with-the-boys/UserService/internal/domain"
)

type JwtAuth interface {
	GenerateTokens(customer *domain.User) (string, string, error)
	ValidateToken(tokenString string) (*domain.User, error)
	RefreshTokens(refreshToken string) (string, string, error)
}
