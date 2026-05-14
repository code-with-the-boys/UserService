package auth_test

import (
	"testing"
	"time"

	"github.com/code-with-the-boys/UserService/internal/domain"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const (
	testAccessSecret  = "test-access-secret-at-least-32-chars-long!!"
	testRefreshSecret = "test-refresh-secret-at-least-32-chars-long"
)

func setup(t *testing.T) auth.JwtAuth {
	t.Helper()
	t.Setenv("JWT_ACCESS_SECRET", testAccessSecret)
	t.Setenv("JWT_REFRESH_SECRET", testRefreshSecret)
	return auth.NewJwtService()
}

func testUser() *domain.User {
	return &domain.User{
		UserID: uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
		Email:  "test@example.com",
	}
}

// parseUnverified извлекает claims без проверки подписи.
// Используется для инспекции сгенерированных токенов в тестах.
func parseUnverified(t *testing.T, tokenString string) jwt.MapClaims {
	t.Helper()
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("ParseUnverified: %v", err)
	}
	return token.Claims.(jwt.MapClaims)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewJwtService_ReadsSecrets(t *testing.T) {
	svc := setup(t)
	// Не экспортировано, но мы можем косвенно проверить, что secrets установлены,
	// сгенерировав и провалидировав токен.
	user := testUser()
	access, _, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}
	_, err = svc.ValidateToken(access)
	if err != nil {
		t.Fatalf("ValidateToken with default secrets failed: %v", err)
	}
}

func TestGenerateTokens_Success(t *testing.T) {
	svc := setup(t)
	user := testUser()

	access, refresh, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("tokens must not be empty")
	}
	if access == refresh {
		t.Fatal("access and refresh tokens must be different")
	}

	// Проверяем claims access-токена
	claims := parseUnverified(t, access)
	if claims["user_id"] != user.UserID.String() {
		t.Errorf("user_id = %v, want %v", claims["user_id"], user.UserID.String())
	}
	if claims["email"] != user.Email {
		t.Errorf("email = %v, want %v", claims["email"], user.Email)
	}
	if claims["token_type"] != "access" {
		t.Errorf("token_type = %v, want access", claims["token_type"])
	}
	if claims["iss"] != "UserService" {
		t.Errorf("iss = %v, want UserService", claims["iss"])
	}

	// Проверяем claims refresh-токена
	claims = parseUnverified(t, refresh)
	if claims["token_type"] != "refresh" {
		t.Errorf("refresh token_type = %v, want refresh", claims["token_type"])
	}
}

func TestGenerateTokens_Expiration(t *testing.T) {
	svc := setup(t)
	user := testUser()

	access, _, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	claims := parseUnverified(t, access)
	expRaw, ok := claims["exp"]
	if !ok {
		t.Fatal("access token missing exp claim")
	}
	exp := time.Unix(int64(expRaw.(float64)), 0)
	expectedExp := time.Now().Add(15 * time.Minute)
	delta := expectedExp.Sub(exp)
	if delta < 0 {
		delta = -delta
	}
	if delta > 5*time.Second {
		t.Errorf("access token exp %v too far from expected %v (delta %v)", exp, expectedExp, delta)
	}
}

func TestValidateToken_ValidAccessToken(t *testing.T) {
	svc := setup(t)
	user := testUser()
	access, _, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	validUser, err := svc.ValidateToken(access)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if validUser.UserID != user.UserID || validUser.Email != user.Email {
		t.Errorf("got %+v, want %+v", validUser, user)
	}
}

func TestValidateToken_ValidRefreshToken(t *testing.T) {
	svc := setup(t)
	user := testUser()
	_, refresh, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	validUser, err := svc.ValidateToken(refresh)
	if err != nil {
		t.Fatalf("ValidateToken on refresh token: %v", err)
	}
	if validUser.UserID != user.UserID {
		t.Errorf("unexpected UserID")
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	svc := setup(t)
	user := testUser()
	access, _, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	// Подменяем последний байт подписи
	corrupted := access[:len(access)-1] + "x"
	_, err = svc.ValidateToken(corrupted)
	if err == nil {
		t.Fatal("expected error for corrupted token")
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	svc := setup(t)
	_, err := svc.ValidateToken("not.a.token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	// Создаём поддельный токен с другим методом (например, HS384)
	user := testUser()
	claims := jwt.MapClaims{
		"user_id":    user.UserID.String(),
		"email":      user.Email,
		"token_type": "access",
		"exp":        time.Now().Add(time.Hour).Unix(),
		"iss":        "UserService",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, err := token.SignedString([]byte(testAccessSecret))
	if err != nil {
		t.Fatalf("failed to sign with HS384: %v", err)
	}

	svc := setup(t)
	_, err = svc.ValidateToken(signed)
	if err == nil {
		t.Fatal("expected error for wrong signing method")
	}
}

func TestValidateToken_MissingTokenType(t *testing.T) {
	user := testUser()
	// Создаём токен без token_type в claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iss":     "UserService",
	})
	signed, err := token.SignedString([]byte(testAccessSecret))
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	svc := setup(t)
	_, err = svc.ValidateToken(signed)
	if err == nil {
		t.Fatal("expected error for token without token_type")
	}
}

func TestValidateToken_UnknownTokenType(t *testing.T) {
	user := testUser()
	claims := jwt.MapClaims{
		"user_id":    user.UserID.String(),
		"email":      user.Email,
		"token_type": "id_token", // неизвестный тип
		"exp":        time.Now().Add(time.Hour).Unix(),
		"iss":        "UserService",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testAccessSecret))
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	svc := setup(t)
	_, err = svc.ValidateToken(signed)
	if err == nil {
		t.Fatal("expected error for unknown token_type")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Для создания просроченного токена мы не можем напрямую через GenerateTokens,
	// поэтому создаём вручную.
	user := testUser()
	claims := jwt.MapClaims{
		"user_id":    user.UserID.String(),
		"email":      user.Email,
		"token_type": "access",
		"exp":        time.Now().Add(-time.Hour).Unix(), // уже истёк
		"iss":        "UserService",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testAccessSecret))
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	svc := setup(t)
	_, err = svc.ValidateToken(signed)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestRefreshTokens_ValidRefreshToken(t *testing.T) {
	svc := setup(t)
	user := testUser()
	access, refresh, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	_ = access // исходный access не используется, только refresh для обновления

	newAccess, newRefresh, err := svc.RefreshTokens(refresh)
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}
	if newAccess == "" || newRefresh == "" {
		t.Fatal("new tokens must not be empty")
	}

	// Проверяем валидность новых токенов и принадлежность пользователю
	validUser, err := svc.ValidateToken(newAccess)
	if err != nil {
		t.Fatalf("validate new access token: %v", err)
	}
	if validUser.UserID != user.UserID {
		t.Errorf("UserID mismatch in refreshed token")
	}

	_, err = svc.ValidateToken(newRefresh)
	if err != nil {
		t.Fatalf("validate new refresh token: %v", err)
	}
}

func TestRefreshTokens_WithAccessTokenFails(t *testing.T) {
	svc := setup(t)
	user := testUser()
	access, _, err := svc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	_, _, err = svc.RefreshTokens(access) // передаём access вместо refresh
	if err == nil {
		t.Fatal("expected error when refreshing with access token")
	}
}

func TestRefreshTokens_InvalidRefreshToken(t *testing.T) {
	svc := setup(t)
	_, _, err := svc.RefreshTokens("invalid.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestRefreshTokens_ExpiredRefreshToken(t *testing.T) {
	// Создаём просроченный refresh токен вручную
	user := testUser()
	claims := jwt.MapClaims{
		"user_id":    user.UserID.String(),
		"email":      user.Email,
		"token_type": "refresh",
		"exp":        time.Now().Add(-time.Hour).Unix(),
		"iss":        "UserService",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testRefreshSecret))
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	svc := setup(t)
	_, _, err = svc.RefreshTokens(signed)
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
}

func TestRefreshTokens_WrongSecret(t *testing.T) {
	// Подписываем refresh токен не тем секретом
	user := testUser()
	claims := jwt.MapClaims{
		"user_id":    user.UserID.String(),
		"email":      user.Email,
		"token_type": "refresh",
		"exp":        time.Now().Add(time.Hour).Unix(),
		"iss":        "UserService",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("wrong-secret-11111111111111111111"))
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	svc := setup(t)
	_, _, err = svc.RefreshTokens(signed)
	if err == nil {
		t.Fatal("expected error for token signed with wrong secret")
	}
}
