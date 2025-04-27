# Start from a small, secure base image for building
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wallet-service ./cmd/app/main.go

# Create a minimal production image
FROM alpine:latest

# Avoid running code as a root user
RUN adduser -D appuser
USER appuser

# Set the working directory inside the container
WORKDIR /app

COPY --from=builder /app/wallet-service .

# Run the binary when the container starts
CMD ["./wallet-service"]
