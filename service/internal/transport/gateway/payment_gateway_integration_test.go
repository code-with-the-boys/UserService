package gateway_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/code-with-the-boys/UserService/internal/domain"
	gw "github.com/code-with-the-boys/UserService/internal/transport/gateway"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const testJWTAccessSecret = "gateway-payment-test-access-secret-minimum-32-chars"
const testJWTRefreshSecret = "gateway-payment-test-refresh-secret-minimum-32-chars"

func testJWT(t *testing.T) auth.JwtAuth {
	t.Helper()
	t.Setenv("JWT_ACCESS_SECRET", testJWTAccessSecret)
	t.Setenv("JWT_REFRESH_SECRET", testJWTRefreshSecret)
	return auth.NewJwtService()
}

func testUser() *domain.User {
	return &domain.User{
		UserID:   uuid.MustParse("00000000-0000-4000-8000-000000000042"),
		Email:    "payment-gateway-e2e@test.local",
		Password: "unused",
	}
}

func newGatewayWithPaymentProxy(t *testing.T, paymentBaseURL string) http.Handler {
	t.Helper()
	logger := zap.NewNop()
	ph, err := gw.NewPaymentHTTPProxy(logger, paymentBaseURL)
	if err != nil {
		t.Fatalf("NewPaymentHTTPProxy: %v", err)
	}
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v1/payment/", http.StripPrefix("/api/v1/payment", ph))
	jwtSvc := testJWT(t)
	return gw.JWTHTTPMiddleware(logger, jwtSvc, mainMux)
}

// Публичный маршрут шлюза → прокси → Payment /health (без JWT).
func TestGatewayPayment_Health_NoJWT(t *testing.T) {
	payment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"status":"ok"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer payment.Close()

	h := newGatewayWithPaymentProxy(t, payment.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/payment/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if body := rr.Body.String(); body != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %q", body)
	}
}

// Защищённый checkout: JWT → шлюз подставляет X-User-Id, Authorization не уходит на Payment.
func TestGatewayPayment_CheckoutSubscription_WithBearer(t *testing.T) {
	var sawUser, sawAuth string
	payment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/checkout/subscription" {
			sawUser = r.Header.Get("X-User-Id")
			sawAuth = r.Header.Get("Authorization")
			w.Header().Set("Location", "https://checkout.stripe.com/c/pay/test-mock")
			w.WriteHeader(http.StatusSeeOther)
			return
		}
		http.NotFound(w, r)
	}))
	defer payment.Close()

	h := newGatewayWithPaymentProxy(t, payment.URL)
	jwtSvc := testJWT(t)
	access, _, err := jwtSvc.GenerateTokens(testUser())
	if err != nil {
		t.Fatal(err)
	}

	body := strings.NewReader(`{"plan_type":"basic"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payment/v1/checkout/subscription", body)
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	want := "00000000-0000-4000-8000-000000000042"
	if sawUser != want {
		t.Fatalf("X-User-Id: got %q want %q", sawUser, want)
	}
	if sawAuth != "" {
		t.Fatalf("Authorization must be stripped on upstream, got %q", sawAuth)
	}
	loc := rr.Result().Header.Get("Location")
	if !strings.Contains(loc, "checkout.stripe.com") {
		t.Fatalf("Location: %q", loc)
	}
}

// Без Bearer на защищённом пути — 401.
func TestGatewayPayment_Checkout_WithoutBearer(t *testing.T) {
	payment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("payment backend must not be called without JWT")
	}))
	defer payment.Close()

	h := newGatewayWithPaymentProxy(t, payment.URL)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payment/v1/checkout/subscription", strings.NewReader(`{"plan_type":"basic"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rr.Code, rr.Body.String())
	}
}

// Webhook публичный — доходит до Payment без JWT.
func TestGatewayPayment_Webhook_NoJWT(t *testing.T) {
	var sawPath string
	payment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/webhooks/stripe" {
			sawPath = r.URL.Path
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"detail":"missing Stripe-Signature header"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer payment.Close()

	h := newGatewayWithPaymentProxy(t, payment.URL)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhooks/stripe", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if sawPath != "/webhooks/stripe" {
		t.Fatalf("upstream path: got %q", sawPath)
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestIsPublicHTTPRoute_PaymentPaths(t *testing.T) {
	if !gw.IsPublicHTTPRoute(http.MethodGet, "/api/v1/payment/health") {
		t.Fatal("payment health should be public")
	}
	if !gw.IsPublicHTTPRoute(http.MethodPost, "/api/v1/payment/webhooks/stripe") {
		t.Fatal("stripe webhook should be public")
	}
	if gw.IsPublicHTTPRoute(http.MethodPost, "/api/v1/payment/v1/checkout/subscription") {
		t.Fatal("checkout must require JWT")
	}
}
