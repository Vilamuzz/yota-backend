# Use a stable Go version (adjust to 1.21 or 1.22 if needed for compatibility)
FROM golang:1.25-alpine AS builder

# Set destination for COPY
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code (including subdirectories)
COPY . ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use a minimal runtime image for the final stage (optional but recommended for security/smaller size)
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .

# Expose the default port (matches your app's default)
EXPOSE 8080

# Run the binary
CMD ["./main"]