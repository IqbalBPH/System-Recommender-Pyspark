FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for module downloads
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Build migrator
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrator ./cmd/migrator

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binaries from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrator .

# Copy migration files
COPY --from=builder /app/migrations ./migrations

# Copy environment file
COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./main"]