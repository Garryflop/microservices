# Order & Payment Microservices

**Course:** Advanced Programming 2  
**Student:** Saparbekov Nurdaulet  
**Group:** SE-2402

Two-service platform in Go following **Clean Architecture** with **gRPC** inter-service communication and **REST** external API (Gin framework).

## Proto Repositories

| Repository | Description |
|-----------|-------------|
| [microservices-proto](https://github.com/Garryflop/microservices-proto) | `.proto` definitions (Contract-First) |
| [microservices-gen-go](https://github.com/Garryflop/microservices-gen-go) | Auto-generated Go code (via GitHub Actions) |

## System Overview

![System Overview](docs/images/Overview.png)

## Architecture Diagram
![Architecture Diagram](docs/images/ContractFirstDraft.png)
![Architecture Diagram](docs/images/ContractFirst.png)

## Clean Architecture Layers

![Clean Architecture Layers](docs/images/Layers.png)

## Order State Diagram

![Order State Diagram](docs/images/OrderState.png)

## Project Structure

```
microservices/
├── order-service/
│   ├── cmd/
│   │   ├── order-service/main.go      # Composition Root (REST + gRPC servers)
│   │   └── stream-client/main.go      # Streaming demo client
│   ├── internal/
│   │   ├── config/config.go           # .env loader
│   │   ├── domain/order.go            # Entity + business rules 
│   │   ├── dto/                       # Request/Response objects
│   │   ├── usecase/                   # Business logic + ports
│   │   ├── repository/               # PostgreSQL persistence
│   │   ├── transport/
│   │   │   ├── http/                  # Gin REST handlers 
│   │   │   └── grpc/server.go         # gRPC streaming server
│   │   ├── infrastructure/
│   │   │   ├── payment_client.go      # OLD HTTP client (kept for reference)
│   │   │   └── grpc_payment_client.go # NEW gRPC client
│   │   └── middleware/                # Logging, recovery, idempotency
│   └── migrations/
├── payment-service/
│   ├── cmd/payment-service/main.go    # Composition Root (REST + gRPC servers)
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── domain/payment.go          # Entity + business rules 
│   │   ├── dto/
│   │   ├── usecase/                   # Business logic 
│   │   ├── repository/               # PostgreSQL persistence
│   │   ├── transport/
│   │   │   ├── http/                  # Gin REST handlers
│   │   │   └── grpc/server.go         # gRPC server
│   │   └── middleware/
│   │       ├── interceptor.go         # gRPC logging interceptors
│   │       ├── logging.go
│   │       └── recovery.go
│   └── migrations/
└── docs/
    └── images/                        # Rendered diagrams + screenshots
```

## Architecture Decisions

| Principle | Implementation |
|-----------|---------------|
| **Clean Architecture** | Domain > Use Case > Repository/Transport. Dependencies point inward. |
| **Contract-First** | Proto definitions in separate repo, code auto-generated via GitHub Actions. |
| **Database per Service** | `orders_db` and `payments_db` - separate databases. |
| **No Shared Code** | Each service has its own models, DTOs, middleware. Zero shared packages. |
| **Manual DI** | All dependencies wired explicitly in `main.go` (Composition Root). |
| **Ports & Adapters** | Use cases depend on interfaces (`OrderRepository`, `PaymentClient`), not implementations. |
| **gRPC for Internal** | Inter-service calls use gRPC. REST kept for external API. |

## What Changed from Assignment 1

The core architecture and business logic remained completely isolated from the gRPC migration, adhering strictly to Clean Architecture principles. The Domain layer (entities and business rules), Use Case layer, Ports (interfaces like `PaymentClient`), and Repository layer (PostgreSQL queries) required absolutely no changes. The modifications were exclusively contained within the outer layers: the Transport layer was updated to include new gRPC and streaming servers, the Infrastructure layer saw the old HTTP client replaced with a new gRPC client, and the Configuration was adjusted to point to gRPC addresses instead of REST URLs.

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
PAYMENT_GRPC_ADDR=localhost:50051
GRPC_PORT=50052
PORT=8081
```

**payment-service/.env**
```env
DB_DSN=postgres://postgres:YOUR_PASSWORD@localhost:5432/payments_db?sslmode=disable
PORT=8082
GRPC_PORT=50051
```

### Step 3: Run Services

```bash
# Terminal 1 - Payment Service (gRPC + REST)
cd payment-service
go run ./cmd/payment-service

# Terminal 2 - Order Service (REST + gRPC streaming)
cd order-service
go run ./cmd/order-service
```

## API Examples

### Create Order

```bash
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{"customer_id": "cust-001", "item_name": "Laptop Stand", "amount": 15000}'
```

Internally, the Order Service calls the Payment Service via **gRPC** (not REST).

### Get Order

```bash
curl http://localhost:8081/orders/{order_id}
```

### Cancel Order

```bash
curl -X PATCH http://localhost:8081/orders/{order_id}/cancel
```

### Subscribe to Order Updates (gRPC Streaming)

```bash
# Terminal 3 - run the streaming client
cd order-service
go run ./cmd/stream-client -order {order_id}
```

Output:
```
subscribed to order {order_id} updates...
status update: Pending (at 12:00:00)
status update: Paid (at 12:00:04)
stream ended
```

## gRPC Services

| Service | Method | Type | Port |
|---------|--------|------|------|
| `PaymentService` | `ProcessPayment` | Unary | `:50051` |
| `OrderTrackingService` | `SubscribeToOrderUpdates` | Server-side Stream | `:50052` |

## gRPC Interceptor (Bonus)

The Payment Service uses a **Unary Logging Interceptor** that logs every incoming gRPC request:

```
gRPC | method=/payment.PaymentService/ProcessPayment | duration=1.2ms | ok
```

## Business Rules

| Rule | Detail |
|------|--------|
| Financial accuracy | `int64` for money (cents). Never `float64`. |
| Amount validation | Must be > 0. |
| Payment limit | Amount > 100,000 cents ($1,000) results in "Declined". |
| Order cancellation | "Paid" orders cannot be cancelled. |
| gRPC error codes | `InvalidArgument` for bad input, `NotFound` for missing orders, `Internal` for server errors. |
| Service unavailable | If Payment Service is down, order is marked "Failed", returns `503`. |
| Idempotency | `Idempotency-Key` header prevents duplicate orders on retry. |

## Failure Handling

When Payment Service is unavailable:
1. Order Service **does not hang** - gRPC client has context-based timeout.
2. Returns **503 Service Unavailable**.
3. Order is marked **"Failed"** - clear signal that the operation did not succeed.

## Streaming Implementation

The `SubscribeToOrderUpdates` stream is tied to **real database changes**:

1. Client subscribes with an `order_id`
2. Server sends the current status immediately
3. Server polls the database every 2 seconds
4. When status changes in DB => server pushes update to the stream
5. Stream ends when a terminal state is reached (Paid / Failed / Cancelled)



---

Made by Harryfloppa with ❤️