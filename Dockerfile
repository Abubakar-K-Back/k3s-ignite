# Build Stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o ignite-brain ./cmd/brain/main.go

# Run Stage
FROM alpine:latest
WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/ignite-brain .

# --- FIX START ---
# Ensure we copy the templates folder into the final image
COPY --from=builder /app/templates ./templates
# --- FIX END ---

EXPOSE 8080
CMD ["./ignite-brain"]