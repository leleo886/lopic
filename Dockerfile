# Build stage
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Set GOPROXY to use domestic mirror
ENV GOPROXY=https://goproxy.cn,direct

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project (embed files need source files during build)
COPY . .

# Build the application with CGO disabled
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/api .

# Create necessary directories
RUN mkdir -p data/uploads data/logs data/backup configs

# Install SQLite (for SQLite database support)
RUN apk add --no-cache sqlite

# Expose the port the app runs on
EXPOSE 6060

# Run the application
# Config file will be created from embedded default if not exists
CMD ["./api", "--serve", "--config=configs/docker_config.yaml"]
