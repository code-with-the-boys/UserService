# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

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

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]