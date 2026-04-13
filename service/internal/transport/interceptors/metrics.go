package interceptors

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	// Total requests
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "code"},
	)

	// Duration histogram
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)

	// In-flight gauge
	requestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_requests_in_flight",
			Help: "Number of requests in flight",
		},
		[]string{"method"},
	)

	// Request size histogram (optional)
	requestSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_size_bytes",
			Help:    "Size of gRPC request messages",
			Buckets: prometheus.ExponentialBuckets(64, 2, 10), // 64,128,256,...,65536
		},
		[]string{"method"},
	)

	// Response size histogram (optional)
	responseSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_response_size_bytes",
			Help:    "Size of gRPC response messages",
			Buckets: prometheus.ExponentialBuckets(64, 2, 10),
		},
		[]string{"method"},
	)
)

func PrometheusInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod

		requestsInFlight.WithLabelValues(method).Inc()
		defer requestsInFlight.WithLabelValues(method).Dec()

		if pb, ok := req.(proto.Message); ok {
			requestSizeBytes.WithLabelValues(method).Observe(float64(proto.Size(pb)))
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if resp != nil {
			if pb, ok := resp.(proto.Message); ok {
				responseSizeBytes.WithLabelValues(method).Observe(float64(proto.Size(pb)))
			}
		}

		code := status.Code(err).String()
		requestsTotal.WithLabelValues(method, code).Inc()
		requestDuration.WithLabelValues(method, code).Observe(duration.Seconds())

		return resp, err
	}
}
