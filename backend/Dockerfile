FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go module files first to leverage build cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
# We're building the API service, but you can modify this for other services like collector
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Add certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /api-server /app/api-server

# Copy .env into root directory for the config loader to find
COPY --from=builder /app/.env /.env

# Create directory for potential data storage
RUN mkdir -p /app/data

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["/app/api-server"]