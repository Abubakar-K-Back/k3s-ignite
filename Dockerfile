FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o brain cmd/brain/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/brain .
CMD ["./brain"]