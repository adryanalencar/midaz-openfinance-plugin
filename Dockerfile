# ==============================================================================
# Multi-Stage Dockerfile for Midaz Open Finance Plugin
# ==============================================================================

# Stage 1: Build binary
FROM golang:1.24-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
# Download dependencies
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/plugin-openfinance \
    cmd/app/main.go

# Stage 2: Minimal Production Image
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user
RUN addgroup -S midaz && adduser -S midaz -G midaz
USER midaz

COPY --from=builder /bin/plugin-openfinance /app/plugin-openfinance
COPY --from=builder /src/migrations /app/migrations
COPY --from=builder /src/.env.example /app/.env.example

EXPOSE 8085

ENTRYPOINT ["/app/plugin-openfinance"]
