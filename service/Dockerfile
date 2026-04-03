# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

ENV PATH="/go/bin:${PATH}"

# go.mod replace points to ../protos — must exist before go mod download (path = /protos when WORKDIR is /app)
COPY protos /protos
COPY service/go.mod service/go.sum ./
RUN go mod download

COPY service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Stage 2: Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash postgresql-client

COPY --from=builder /go/bin/goose /usr/local/bin/

COPY --from=builder /app/main /app/

COPY --from=builder /app/internal/db/migrations /app/migrations

COPY --from=builder /app/internal/scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

WORKDIR /app

EXPOSE 8080 9090

ENTRYPOINT ["/app/entrypoint.sh"]