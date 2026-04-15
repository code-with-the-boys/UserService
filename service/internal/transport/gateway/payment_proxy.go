package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	customContext "github.com/code-with-the-boys/UserService/internal/context"
	"go.uber.org/zap"
)

// NewPaymentHTTPProxy returns a handler that forwards to PaymentService (Litestar) after StripPrefix.
// Register as: mainMux.Handle("/api/v1/payment/", http.StripPrefix("/api/v1/payment", h)).
// Injects X-User-Id from JWT context when present; strips Authorization so the token does not leak downstream.
func NewPaymentHTTPProxy(logger *zap.Logger, paymentBaseURL string) (http.Handler, error) {
	base := strings.TrimSpace(paymentBaseURL)
	if base == "" {
		return nil, fmt.Errorf("empty PAYMENT_SERVICE_HTTP_URL")
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("payment URL must include scheme and host (e.g. http://payment-service:8000)")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.Host = u.Host
			req.Header.Del("Authorization")
			if auth, err := customContext.GetAuth(req.Context()); err == nil && auth.User != nil {
				req.Header.Set("X-User-Id", auth.User.UserID.String())
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if logger != nil {
				logger.Error("payment proxy error", zap.String("path", r.URL.Path), zap.Error(err))
			}
			http.Error(w, `{"error":"payment service unreachable"}`, http.StatusBadGateway)
		},
	}
	return proxy, nil
}
