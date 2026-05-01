FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service ./cmd/order-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /order-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8081 50052

CMD ["./order-service"]
