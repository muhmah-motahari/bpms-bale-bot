# ---- Build Stage ----
FROM golang:1.23 as builder
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o app ./cmd/main

# ---- Run Stage ----
FROM debian:bookworm-slim
WORKDIR /app

# Copy the built binary from the builder
COPY --from=builder /app/app ./app

# Copy the environment file
COPY .env .env

# EXPOSE 8080

# Run the binary
ENTRYPOINT ["./app"] 