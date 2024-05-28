DO $$ BEGIN -- Создание схемы, если она не существует
IF NOT EXISTS (
    SELECT 1
    FROM pg_namespace
    WHERE nspname = 'ecommerce'
) THEN EXECUTE 'CREATE SCHEMA ecommerce AUTHORIZATION pg_database_owner';
END IF;
-- Таблица deliveries
IF NOT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_tables
    WHERE schemaname = 'ecommerce'
        AND tablename = 'deliveries'
) THEN CREATE TABLE ecommerce.deliveries (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL CHECK (phone ~ '^\+\d{1,15}$'),
    zip TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    region TEXT NOT NULL,
    email TEXT NOT NULL CHECK (
        email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
    )
);
END IF;
-- Таблица payments
IF NOT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_tables
    WHERE schemaname = 'ecommerce'
        AND tablename = 'payments'
) THEN CREATE TABLE ecommerce.payments (
    id UUID PRIMARY KEY,
    transaction TEXT NOT NULL UNIQUE,
    request_id TEXT NOT NULL,
    currency TEXT NOT NULL CHECK (currency IN ('USD', 'EUR', 'RUB')),
    provider TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    payment_dt BIGINT NOT NULL,
    bank TEXT NOT NULL,
    delivery_cost BIGINT NOT NULL CHECK (delivery_cost >= 0),
    goods_total BIGINT NOT NULL CHECK (goods_total > 0),
    custom_fee BIGINT NOT NULL CHECK (custom_fee >= 0)
);
END IF;
-- Таблица orders
IF NOT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_tables
    WHERE schemaname = 'ecommerce'
        AND tablename = 'orders'
) THEN CREATE TABLE ecommerce.orders (
    order_uid UUID PRIMARY KEY,
    track_number UUID NOT NULL UNIQUE,
    entry TEXT,
    delivery_id UUID REFERENCES ecommerce.deliveries(id) ON DELETE CASCADE,
    payment_id UUID REFERENCES ecommerce.payments(id) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    internal_signature TEXT,
    customer_id UUID NOT NULL,
    delivery_service TEXT NOT NULL,
    shardkey TEXT NOT NULL,
    sm_id BIGINT NOT NULL,
    date_created TIMESTAMP NOT NULL DEFAULT NOW(),
    oof_shard TEXT
);
END IF;
-- Таблица items
IF NOT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_tables
    WHERE schemaname = 'ecommerce'
        AND tablename = 'items'
) THEN CREATE TABLE ecommerce.items (
    id UUID PRIMARY KEY,
    chrt_id BIGINT NOT NULL,
    track_number TEXT NOT NULL,
    price BIGINT NOT NULL,
    rid TEXT NOT NULL,
    name TEXT NOT NULL,
    sale INT NOT NULL,
    size TEXT NOT NULL,
    total_price BIGINT NOT NULL,
    nm_id BIGINT NOT NULL,
    brand TEXT NOT NULL,
    status INT NOT NULL
);
END IF;
END $$;
