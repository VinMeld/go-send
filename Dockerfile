# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o go-send-server cmd/server/main.go

# Final Stage
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/go-send-server .

# Expose port
EXPOSE 8080

# Create volume for local storage
VOLUME ["/root/server_data"]

# Run the server
CMD ["./go-send-server", "-port", ":8080"]
