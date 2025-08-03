# Go Micro Payment System

 Microservices-based e-commerce payment system built with Go. It leverages gRPC for synchronous communication, Kafka for asynchronous event-driven messaging, and PostgreSQL for data persistence. The system is designed to be scalable, fault-tolerant, and maintainable by decomposing business logic into independent services.

<p align="center">
    <img src="./docs/assets/wall.webp" alt="wall" width="450">
</p>

---

## System Architecture

The architecture consists of a set of distinct microservices, each with a specific domain responsibility. An **API Gateway** serves as the single entry point for all client requests, routing them to the appropriate backend service.

The core services include:
* **User Service**: Manages user identity, including account creation and authentication.
* **Product Service**: Responsible for the product catalog, including details like name, description, and price.
* **Inventory Service**: Manages stock levels for all products, handling reservations and adjustments.
* **Order Service**: Orchestrates the order creation and management process, interacting with multiple other services.
* **Payment Service**: Handles all payment-related logic, including integration with external payment processors like Stripe.
* **Notification Service**: Manages the sending of user notifications, such as email or SMS.

Each service has its own dedicated PostgreSQL database to ensure loose coupling and independent data management.

---

## Communication Patterns

The system employs a dual communication strategy to handle different interaction requirements effectively.

### gRPC for Synchronous Communication
For operations requiring an immediate response, the system uses **gRPC**. This high-performance RPC framework is ideal for low-latency, request-response communication between services. The use of Protocol Buffers ensures strongly-typed contracts and efficient data serialization. This pattern is primarily used when the **API Gateway** calls internal services or when a service like **Order** needs to synchronously orchestrate calls to the **Product**, **Inventory**, and **Payment** services to validate and create an order.

### Kafka for Asynchronous Event-Driven Communication
To ensure decoupling, resilience, and scalability, **Kafka** is used for asynchronous messaging. When a significant business event occurs (e.g., an order is created, or a payment is processed), the responsible service publishes an event to a specific Kafka topic. Other services subscribe to these topics and react to the events accordingly. This event-driven approach allows services to operate independently and process events at their own pace. For instance, the **Notification Service** consumes `order.created` events to send confirmations, and the **Order Service** listens for `payment.status-updated` events to finalize an order's state. This decouples the order creation logic from the notification and payment finalization processes.

---

## Interaction Flow Example: Order Creation

1.  A client sends a JSON request to create an order to the **API Gateway**.
2.  The Gateway validates the request, converts the JSON payload to a Protobuf message, and makes a synchronous **gRPC** call to the `CreateOrder` endpoint on the **Order Service**.
3.  The **Order Service** begins orchestrating the request by making sequential, synchronous **gRPC** calls to:
    * The **Product Service** to fetch product details and prices.
    * The **Inventory Service** to check for stock availability and then to reserve the required stock.
    * The **Payment Service** to initiate a payment transaction.
4.  Once these synchronous calls succeed, the Order Service persists the order with a `PENDING` status in its database and publishes an `order.created` event to a **Kafka** topic.
5.  Asynchronously, downstream services consume this event:
    * The **Notification Service** consumes the `order.created` event and sends a confirmation email to the user.
    * The **Payment Service**, after receiving a webhook from Stripe, publishes a `payment.status-updated` event.
    * The **Order Service** consumes the `payment.status-updated` event to update the order's status from `PENDING` to `PAID` or `FAILED`, thus completing the lifecycle.

---

## Technology Stack

* **Language**: Go
* **Database**: PostgreSQL (with GORM)
* **Synchronous Communication**: gRPC
* **Asynchronous Communication**: Kafka (with the Sarama client)
* **External Services**: Stripe for payments, SMTP for email notifications

---

## Setup Instructions

1.  **Prerequisites**: Ensure Go 1.21+, PostgreSQL, Kafka, and the `protoc` compiler are installed and accessible in your environment. Running PostgreSQL and Kafka via Docker is recommended.
2.  **Clone Repository**:
    ```bash
    git clone [https://github.com/SabinGhost19/go-micro-payment](https://github.com/SabinGhost19/go-micro-payment)
    cd go-micro-payment
    ```
3.  **Install Dependencies**:
    ```bash
    go mod tidy
    ```
4.  **Generate gRPC Code**: Compile the `.proto` files to generate the necessary Go code for gRPC.
    ```bash
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           proto/*/*.proto
    ```
5.  **Environment Configuration**: Create a `.env` file within each service's directory (e.g., `services/order/.env`). Populate it with the required environment variables, such as database connection strings (`DB_DSN`) and service addresses.
6.  **Run Services**: Compile and run each microservice in a separate terminal session. For example, to run the user service:
    ```bash
    go build -o user-service ./services/user && ./user-service
    ```
    Repeat this command for all services (product, inventory, order, payment, notification, and api-gateway)



