# ---- Build Stage ----
FROM golang:1.23 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/main

# ---- Run Stage ----
FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app ./app
COPY wait-for-it.sh ./wait-for-it.sh
COPY .env .env

RUN chmod +x ./wait-for-it.sh

ENTRYPOINT ["./wait-for-it.sh", "db:5432", "--", "./app"] 