# Pay-As-You-Use Billing System

![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=flat&logo=postgresql&logoColor=white)
![Stripe](https://img.shields.io/badge/Stripe-Webhooks-635BFF?style=flat&logo=stripe&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker&logoColor=white)

A production-grade microservices billing engine built in Go. Designed for API-driven platforms and AI SaaS products where users purchase token packages and consume credits synchronously per action.

---

## Architecture & Engineering Highlights

* **Clean & Maintainable Architecture:** Clear separation of concerns across transport (`handler`), business logic (`service`), domain models (`domain`), and data access (`repository`) layers.
* **Append-Only Immutable Ledger:** User balances are never stored as a mutable integer column. Instead, balances are calculated dynamically from an immutable transaction log (`payment_schema.ledger`), preventing race conditions and providing a complete financial audit trail.
* **Saga-Like Distributed Transactions:** When a user triggers a paid action in the `business-service`, a `pending` usage log is created first. The service then makes a synchronous HTTP call to `payment-service` with a strict **3-second context timeout**. Based on the deduction outcome, the usage log transitions to `success` or `failed`.
* **Idempotent Stripe Webhooks:** Asynchronous payment processing using `checkout.session.completed` and `checkout.session.expired` events, tracking full state history in `deposit_status_history`.

---

## System Flow (Deposit & Webhook Checkout)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant Payment as Payment Service (:8000)
    participant DB_Pay as PostgreSQL (payment_schema)
    participant Stripe as Stripe API

    Client->>Payment: POST /api/v1/checkout/:user_id (package_id)
    Payment->>DB_Pay: Verify package & INSERT deposit_request
    DB_Pay-->>Payment: request_id (UUID)
    Payment->>DB_Pay: INSERT deposit_status_history (status = 'pending')
    
    Payment->>Stripe: Create Checkout Session (metadata: request_id)
    Stripe-->>Payment: checkout_url & session_id
    Payment->>DB_Pay: UPDATE deposit_request (stripe_session_id)
    Payment-->>Client: 200 OK (checkout_url)

    Note over Client,Stripe: User completes payment on Stripe Checkout page

    Stripe->>Payment: POST /api/v1/webhook (checkout.session.completed)
    Payment->>Payment: Verify Stripe-Signature & extract request_id
    Payment->>DB_Pay: Begin Transaction
    Payment->>DB_Pay: INSERT deposit_status_history (status = 'completed')
    Payment->>DB_Pay: INSERT ledger (+tokens, ref: request_id, type: 'deposit')
    DB_Pay-->>Payment: Commit Transaction
    Payment-->>Stripe: 200 OK
```

## System Flow (Token Consumption)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant Business as Business Service (:8001)
    participant DB_Biz as PostgreSQL (business_schema)
    participant Payment as Payment Service (:8000)
    participant DB_Pay as PostgreSQL (payment_schema)

    Client->>Business: POST /api/v1/business/consume/:user_id
    Business->>DB_Biz: INSERT usage_log (status = 'pending')
    DB_Biz-->>Business: log_id (UUID)
    
    Business->>Payment: POST /api/v1/deduct (timeout: 3s, ref: log_id)
    Payment->>DB_Pay: Begin Transaction & Check Balance
    alt Balance >= Cost
        Payment->>DB_Pay: INSERT ledger (-amount, ref: log_id)
        DB_Pay-->>Payment: Commit
        Payment-->>Business: 200 OK
        Business->>DB_Biz: UPDATE usage_log (status = 'success')
        Business-->>Client: 200 OK (Action Executed)
    else Insufficient Funds / Timeout
        Payment-->>Business: 402 Payment Required / Error
        Business->>DB_Biz: UPDATE usage_log (status = 'failed')
        Business-->>Client: 402 Payment Required
    end
```

---

## Project Structure

```text
.
├── cmd
│   ├── business/main.go       # Entrypoint for Business Service (:8001)
│   ├── payment/main.go        # Entrypoint for Payment & Ledger Service (:8000)
│   └── user/main.go           # Entrypoint for User Registry Service (:8002)
├── internal
│   ├── client/payment.go      # HTTP client with context timeout for inter-service calls
│   ├── database/db.go         # pgxpool connection management
│   ├── domain/                # Core entities, constants, and domain errors
│   ├── handler/               # Fiber HTTP handlers & Stripe webhook verification
│   ├── repository/            # Raw SQL repositories per schema
│   └── service/               # Business logic & transaction orchestration
├── migrations/                # SQL schema migrations (golang-migrate)
├── Dockerfile.business
├── Dockerfile.payment
├── Dockerfile.user
└── docker-compose.yaml
```

---

## Quick Start

### 1. Prerequisites
* **Docker** & **Docker Compose**
* **golang-migrate** CLI (for applying SQL migrations)
* **Stripe CLI** (optional, for testing webhooks locally)

### 2. Environment Configuration
Create a `.env` file in the root directory and fill it as it showed in `.env.example`:

```env
PAYMENT_PORT=8000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secretpassword
DB_NAME=billing_db

STRIPE_SECRET_KEY= 
STRIPE_WEBHOOK_SECRET=

BASE_URL=http://127.0.0.1:8000
FRONTEND_URL=http://127.0.0.1:3000
PAYMENT_URL=http://payment-api:8000

BUSINESS_PORT=8001
USER_PORT=8002
```

### 3. Build and Run Services
Start PostgreSQL and all three microservices in detached mode:

```bash
docker compose up --build -d
```

### 4. Apply Database Migrations
Run the migrations against the running PostgreSQL container:

```bash
migrate -path migrations -database "postgres://postgres:secretpassword@localhost:5432/billing_db?sslmode=disable" up
```

---
