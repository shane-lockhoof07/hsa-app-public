# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o hsa-api main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/hsa-api .

# Copy migrations directory
COPY migrations ./migrations

# Create data directory
RUN mkdir -p /data/hsa

# Expose port
EXPOSE 8080

# Run the application
CMD ["./hsa-api"]