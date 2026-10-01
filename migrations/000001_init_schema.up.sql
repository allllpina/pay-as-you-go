-- Додаємо розширення для генерації UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Створюємо ізольовані схеми
CREATE SCHEMA user_schema;
CREATE SCHEMA payment_schema;
CREATE SCHEMA business_schema;

-- Таблиці для домену Payment
CREATE TABLE payment_schema.token_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stripe_price_id VARCHAR(100) UNIQUE NOT NULL,
    tokens_amount INTEGER NOT NULL,
    price_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE payment_schema.deposit_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    stripe_session_id VARCHAR(255) UNIQUE,
    package_id UUID REFERENCES payment_schema.token_packages(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payment_schema.deposit_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deposit_request_id UUID REFERENCES payment_schema.deposit_requests(id),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payment_schema.ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    amount INTEGER NOT NULL,
    reference_id UUID,
    operation_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
