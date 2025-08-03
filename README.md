# Go Micro Payment System

A microservices-based e-commerce payment system built with Go, gRPC, Kafka, and PostgreSQL. The system manages user accounts, product catalogs, inventory, orders, payments, and notifications, with an API Gateway as the client entry point.

## Architecture Overview

The system consists of six microservices: **User**, **Product**, **Inventory**, **Order**, **Payment**, **Notification**, and **API Gateway**. Here's how they interact and their roles:

### User Service
- **Purpose**: Manages user account creation, authentication, and profile queries.
- **gRPC Role**: Acts as a gRPC server for `CreateUser`, `GetUser`, and `AuthenticateUser` endpoints. No gRPC client role, as it doesn't call other services directly.
- **Kafka Role**: Publishes `user.created` events to Kafka when a user is created. No consumer role.
- **Database**: Stores user records (PostgreSQL).

### Product Service
- **Purpose**: Manages the product catalog (name, description, price, stock).
- **gRPC Role**: Acts as a gRPC server for `CreateProduct`, `GetProduct`, `ListProducts`, `UpdateProduct`, and `DeleteProduct` endpoints. Calls the Inventory Service's `UpdateStock` endpoint for stock updates.
- **Kafka Role**: Publishes `product.created`, `product.updated`, and `product.deleted` events to Kafka for inventory synchronization.
- **Database**: Stores product records (PostgreSQL).

### Inventory Service
- **Purpose**: Manages stock levels and reservations for products.
- **gRPC Role**: Acts as a gRPC server for `CheckStock`, `ReserveStock`, and `UpdateStock` endpoints. Calls the Product Service's `GetProduct` endpoint to validate products.
- **Kafka Role**: Publishes `stock.reserved` and `stock.updated` events to Kafka. Consumes `product.created`, `product.updated`, and `product.deleted` events to sync inventory.
- **Database**: Stores inventory records (PostgreSQL).

### Order Service
- **Purpose**: Manages order creation, status updates, and queries.
- **gRPC Role**: Acts as a gRPC server for `CreateOrder`, `GetOrder`, and `ListOrders` endpoints. Acts as a gRPC client when calling the Product Service (`GetProduct`), Inventory Service (`CheckStock`, `ReserveStock`), and Payment Service (`InitiatePayment`).
- **Kafka Role**: Publishes `order.created` events to Kafka when an order is created. Consumes `payment.status-updated` and `stock-events` to update order status (e.g., from PENDING to PAID or FAILED).
- **Database**: Stores orders and order items (PostgreSQL).

### Payment Service
- **Purpose**: Handles payment processing via Stripe and updates payment status.
- **gRPC Role**: Acts as a gRPC server for `InitiatePayment` and `CheckPaymentStatus` endpoints. No gRPC client role.
- **Kafka Role**: Publishes `payment.created` and `payment.status-updated` events to Kafka. Listens to Stripe webhooks to update payment status and publishes updates to Kafka.
- **Database**: Stores payment records (PostgreSQL).

### Notification Service
- **Purpose**: Sends email or SMS notifications to users.
- **gRPC Role**: Acts as a gRPC server for `SendEmail` and `SendSMS` endpoints. No gRPC client role.
- **Kafka Role**: Consumes `order.created` and `payment.status-updated` events to send notifications (e.g., order confirmation, payment status). Publishes `notification.sent` events for logging/audit.
- **Database**: Stores notification records (PostgreSQL).

### API Gateway
- **Purpose**: Acts as the entry point for external clients (e.g., web/mobile apps). Converts JSON requests to Protobuf and routes them to the appropriate gRPC service.
- **gRPC Role**: Acts as a gRPC client, calling User, Product, Inventory, Order, Payment, or Notification services based on the request.
- **Kafka Role**: No direct Kafka interaction, as it focuses on request routing.
- **Database**: None (stateless).

## Communication Patterns
- **gRPC**: Used for synchronous communication where immediate responses are needed (e.g., creating an order, fetching product details, initiating a payment). The API Gateway calls gRPC endpoints on the services, and the Order Service calls the Product, Inventory, and Payment Services via gRPC.
- **Kafka (Sarama)**: Used for asynchronous event-driven communication. Services publish events to Kafka topics when significant actions occur (e.g., order created, payment status updated, product updated). Other services subscribe to these topics to react (e.g., Notification Service sends emails, Inventory Service syncs stock).
- **Why Kafka?**: Ensures decoupled, scalable, and fault-tolerant communication. If a service is down, it can process missed events later by consuming from Kafka. Supports event sourcing and auditing.
- **Why Sarama?**: A mature Go client for Kafka, offering high performance and reliability with features like consumer groups and offset management.

## Interaction Flow
1. **Client → API Gateway**:
   - Client sends a JSON request (e.g., `POST /orders` to create an order).
   - API Gateway converts JSON to Protobuf and calls the appropriate gRPC endpoint (e.g., Order Service's `CreateOrder`).

2. **Order Service → Product & Inventory Services**:
   - Order Service calls Product Service's `GetProduct` to fetch product prices and Inventory Service's `CheckStock` to verify stock availability.
   - Calls Inventory Service's `ReserveStock` to reserve stock for the order.

3. **Order Service → Payment Service**:
   - Order Service calls Payment Service's `InitiatePayment` to start the payment process.
   - Saves the order (PENDING) and publishes an `order.created` event to Kafka.

4. **Payment Service → Kafka**:
   - Processes payment via Stripe, saves the payment record, and publishes a `payment.created` event.
   - On Stripe webhook, updates payment status and publishes a `payment.status-updated` event.

5. **Notification Service ← Kafka**:
   - Consumes `order.created` and `payment.status-updated` events.
   - Sends email/SMS and publishes a `notification.sent` event.

6. **Order Service ← Kafka**:
   - Consumes `payment.status-updated` and `stock-events` to update order status (e.g., to PAID or FAILED).

7. **Product Service → Inventory Service**:
   - Publishes `product.created`, `product.updated`, and `product.deleted` events to Kafka.
   - Inventory Service consumes these events to sync product stock.

## Kafka Topics
- `user-events`: For `user.created` events.
- `product-events`: For `product.created`, `product.updated`, `product.deleted` events.
- `stock-events`: For `stock.reserved`, `stock.updated` events.
- `order-events`: For `order.created` events.
- `payment-events`: For `payment.created` events.
- `payment-status-updates`: For `payment.status-updated` events.
- `notification-events`: For `notification.sent` events.

## Tech Stack
- **Language**: Go
- **Database**: PostgreSQL (via GORM)
- **Communication**: gRPC (synchronous), Kafka with Sarama (asynchronous)
- **External Services**: Stripe for payments, SMTP for email notifications
- **Dependencies**:
  - `google.golang.org/grpc`
  - `github.com/IBM/sarama`
  - `gorm.io/gorm`
  - `gorm.io/driver/postgres`
  - `github.com/google/uuid`

## Setup Instructions
1. **Prerequisites**:
   - Install Go 1.21+: `https://golang.org/dl/`
   - Install PostgreSQL: `docker run -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=secret -p 5432:5432 postgres`
   - Install Kafka: `docker run -p 9092:9092 confluentinc/cp-kafka`
   - Install `protoc` for gRPC: `sudo apt-get install protobuf-compiler` (or equivalent).

2. **Clone the Repository**:
   ```bash
   git clone github.com/SabinGhost19/go-micro-payment
   cd go-micro-payment
   ```

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

4. **Generate gRPC Files**:
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          proto/*/*.proto
   ```

5. **Set Environment Variables**:
   Create a `.env` file for each service in its directory (`services/<service>`):
   ```bash
   # Example for Order Service (services/order/.env)
   DB_DSN=host=localhost user=admin password=secret dbname=orders port=5432 sslmode=disable
   KAFKA_BROKERS=kafka:9092
   GRPC_PORT=:50051
   PAYMENT_SERVICE_ADDR=payment-service:50052
   INVENTORY_SERVICE_ADDR=inventory-service:50054
   PRODUCT_SERVICE_ADDR=product-service:50055
   ```

6. **Run Services**:
   ```bash
   go build -o user-service ./services/user && ./user-service
   go build -o product-service ./services/product && ./product-service
   go build -o inventory-service ./services/inventory && ./inventory-service
   go build -o order-service ./services/order && ./order-service
   go build -o payment-service ./services/payment && ./payment-service
   go build -o notification-service ./services/notification && ./notification-service
   go build -o api-gateway ./services/api-gateway && ./api-gateway
   ```

## Testing
1. **Create a User**:
   ```bash
   grpcurl -plaintext -d '{"name":"John Doe","email":"john@example.com"}' localhost:50056 userpb.UserService/CreateUser
   ```

2. **Create a Product**:
   ```bash
   grpcurl -plaintext -d '{"name":"Laptop","description":"High-end laptop","price":999.99,"stock":100}' localhost:50055 productpb.ProductService/CreateProduct
   ```

3. **Create an Order**:
   ```bash
   grpcurl -plaintext -d '{"user_id":"USER_UUID","items":[{"product_id":"PRODUCT_UUID","quantity":2}],"address":"123 Main St","currency":"USD"}' localhost:50051 orderpb.OrderService/CreateOrder
   ```

4. **Verify**:
   - Check PostgreSQL tables (`users`, `products`, `inventory`, `orders`, `order_items`, `payments`, `notifications`).
   - Monitor Kafka topics (`user-events`, `product-events`, `stock-events`, `order-events`, `payment-events`, `payment-status-updates`, `notification-events`) using a tool like `kafkacat` or Confluent Control Center.
   - Confirm notifications (email/SMS) are sent.

## Development Notes
- **Error Handling**: Uses gRPC status codes for client errors and transactions for database operations.
- **Logging**: Uses `log.Printf` (replace with a structured logger like Zap in production).
- **Scalability**: Kafka and gRPC ensure decoupled, scalable communication. Use Docker Compose for orchestration in production.
- **Future Improvements**:
  - Add retries for gRPC and Kafka operations.
  - Implement circuit breakers for gRPC calls.
  - Add monitoring (e.g., Prometheus, Grafana).
  - Use compensating transactions for failure scenarios (e.g., release stock if payment fails).