# Build stage
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy Go modules manifests
COPY go.mod ./
RUN go mod tidy

# Copy source code
COPY main.go ./

# Build binary
RUN go build -o mtls-client main.go

# Run stage
FROM alpine:3.18

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mtls-client .

# Expose port
EXPOSE 8443

# Run the server
CMD ["./mtls-client"]
