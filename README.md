# Order & Payment Microservices

**Course:** Advanced Programming 2  
**Student:** Saparbekov Nurdaulet  
**Group:** SE-2402

Two-service platform in Go following **Clean Architecture** with REST communication (Gin framework).

## System Overview

![System Overview](docs/images/Overview.png)

## Project Structure

```
microservices/
├── order-service/                    # Manages orders and their lifecycle
│   ├── cmd/order-service/main.go     # Composition Root (Manual DI)
│   ├── internal/
│   │   ├── config/config.go          # .env loader
│   │   ├── domain/order.go           # Entity + business rules
│   │   ├── dto/                      # Request/Response objects
│   │   ├── usecase/                  # Business logic + port interfaces
│   │   ├── repository/              # PostgreSQL persistence
│   │   ├── transport/http/           # Gin handlers + router
│   │   ├── middleware/               # Logging, recovery, idempotency
│   │   └── infrastructure/          # Payment Service HTTP client
│   └── migrations/
├── payment-service/                  # Processes payments
│   ├── cmd/payment-service/main.go   # Composition Root (Manual DI)
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── domain/payment.go
│   │   ├── dto/
│   │   ├── usecase/
│   │   ├── repository/
│   │   ├── transport/http/
│   │   └── middleware/
│   └── migrations/
└── docs/
    └─images/                         #images for README.md
```

## Architecture Decisions

| Principle | Implementation |
|-----------|---------------|
| **Clean Architecture** | Domain > Use Case > Repository/Transport. Dependencies point inward. |
| **Database per Service** | `orders_db` (Order Service) and `payments_db` (Payment Service) -- separate databases. |
| **No Shared Code** | Each service has its own models, DTOs, middleware. Zero shared packages. |
| **Manual DI** | All dependencies wired explicitly in `main.go` (Composition Root). |
| **Ports & Adapters** | Use cases depend on interfaces (`OrderRepository`, `PaymentClient`), not implementations. |

### Clean Architecture Layers

![Clean Architecture Layers](docs/images/Layers.png)

### Order State Diagram

![Order State Diagram](docs/images/OrderState.png)

## Installation

### Prerequisites

- **Go 1.25+** 
- **PostgreSQL** 

### Step 1: Create Databases

```sql
CREATE DATABASE orders_db;
CREATE DATABASE payments_db;
```

### Step 2: Configure Environment

Copy `.env.example` to `.env` in each service and set your PostgreSQL password:

**order-service/.env**
```env
DB_DSN=postgres://postgres:YOUR_PASSWORD@localhost:5432/orders_db?sslmode=disable
PAYMENT_SERVICE_URL=http://localhost:8082
PORT=8081
```

**payment-service/.env**
```env
DB_DSN=postgres://postgres:YOUR_PASSWORD@localhost:5432/payments_db?sslmode=disable
PORT=8082
```

### Step 3: Run Services

```bash
# Terminal 1
cd payment-service
go run ./cmd/payment-service

# Terminal 2
cd order-service
go run ./cmd/order-service
```

## Business Logic

### Order Service

- **Create Order** (`POST /orders`): Creates order as "Pending", calls Payment Service, then updates to "Paid" or "Failed".
- **Get Order** (`GET /orders/:id`): Returns order details from database.
- **Cancel Order** (`PATCH /orders/:id/cancel`): Only "Pending" orders can be cancelled. "Paid" orders cannot be cancelled.

### Payment Service

- **Process Payment** (`POST /payments`): Validates amount, applies limit rule, stores transaction.
- **Get Payment** (`GET /payments/:order_id`): Returns payment details for a given order.

### Business Rules

| Rule | Detail |
|------|--------|
| Financial accuracy | `int64` for money (cents). Never `float64`. |
| Amount validation | Must be > 0. |
| Payment limit | Amount > 100,000 cents ($1,000) results in "Declined". |
| Order cancellation | "Paid" orders cannot be cancelled. |
| HTTP timeout | Order Service uses 2-second timeout for Payment Service calls. |
| Service unavailable | If Payment Service is down, order is marked "Failed", returns `503`. |
| Idempotency (bonus) | `Idempotency-Key` header prevents duplicate orders on retry. |

### Failure Handling

When Payment Service is unavailable:
1. Order Service **does not hang** -- 2-second HTTP client timeout.
2. Returns **503 Service Unavailable**.
3. Order is marked **"Failed"** (not "Pending") -- clear signal that the operation did not succeed. Client can retry safely.

## API Examples

### Create Order (Authorized)

```bash
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{"customer_id": "cust-001", "item_name": "Laptop Stand", "amount": 15000}'
```

```json
{
  "id": "uuid",
  "customer_id": "cust-001",
  "item_name": "Laptop Stand",
  "amount": 15000,
  "status": "Paid",
  "created_at": "2026-04-01T10:00:00Z"
}
```

### Create Order (Declined — exceeds limit)

```bash
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust-001", "item_name": "Diamond Ring", "amount": 200000}'
```

```json
{
  "id": "uuid",
  "customer_id": "cust-001",
  "item_name": "Diamond Ring",
  "amount": 200000,
  "status": "Failed",
  "created_at": "2026-04-01T10:00:00Z"
}
```

### Get Order

```bash
curl http://localhost:8081/orders/{order_id}
```

### Cancel Order

```bash
curl -X PATCH http://localhost:8081/orders/{order_id}/cancel
```

### Get Payment

```bash
curl http://localhost:8082/payments/{order_id}
```

### Failure Scenario (Payment Service Down)

```bash
# Stop Payment Service (Ctrl+C), then:
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust-001", "item_name": "Test", "amount": 5000}'
```

```json
{"error": "payment service unavailable"}
```


Made by Harryfloppa with ❤️