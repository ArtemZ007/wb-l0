DO $$
BEGIN
    -- Создание схемы, если она не существует
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'ecommerce') THEN
        EXECUTE 'CREATE SCHEMA ecommerce AUTHORIZATION pg_database_owner';
    END IF;

    -- Таблица deliveries
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_tables WHERE schemaname = 'ecommerce' AND tablename = 'deliveries') THEN
        CREATE TABLE ecommerce.deliveries (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            phone TEXT NOT NULL CHECK (phone ~ '^\+\d{1,15}$'),
            zip TEXT NOT NULL,
            city TEXT NOT NULL,
            address TEXT NOT NULL,
            region TEXT NOT NULL,
            email TEXT NOT NULL CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
        );
    END IF;

    -- Таблица payments
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_tables WHERE schemaname = 'ecommerce' AND tablename = 'payments') THEN
        CREATE TABLE ecommerce.payments (
            id SERIAL PRIMARY KEY,
            transaction TEXT NOT NULL UNIQUE,
            request_id TEXT NOT NULL,
            currency TEXT NOT NULL CHECK (currency IN ('USD', 'EUR', 'RUB')),
            provider TEXT NOT NULL,
            amount INT NOT NULL CHECK (amount > 0),
            payment_dt INT NOT NULL,
            bank TEXT NOT NULL,
            delivery_cost INT NOT NULL CHECK (delivery_cost >= 0),
            goods_total INT NOT NULL CHECK (goods_total > 0),
            custom_fee INT NOT NULL CHECK (custom_fee >= 0)
        );
    END IF;

    -- Таблица items
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_tables WHERE schemaname = 'ecommerce' AND tablename = 'items') THEN
        CREATE TABLE ecommerce.items (
            id SERIAL PRIMARY KEY,
            chrt_id INT NOT NULL,
            track_number TEXT NOT NULL UNIQUE,
            price INT NOT NULL CHECK (price > 0),
            rid TEXT NOT NULL,
            name TEXT NOT NULL,
            sale INT NOT NULL CHECK (sale >= 0),
            size TEXT NOT NULL,
            total_price INT NOT NULL CHECK (total_price > 0),
            nm_id INT NOT NULL,
            brand TEXT NOT NULL,
            status INT NOT NULL CHECK (status BETWEEN 1 AND 5)
        );
    END IF;

    -- Таблица orders
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_tables WHERE schemaname = 'ecommerce' AND tablename = 'orders') THEN
        CREATE TABLE ecommerce.orders (
            order_uid UUID PRIMARY KEY,
            track_number UUID NOT NULL UNIQUE,
            entry TEXT,
            delivery_id INT REFERENCES ecommerce.deliveries(id) ON DELETE SET NULL,
            payment_id INT REFERENCES ecommerce.payments(id) ON DELETE SET NULL,
            locale TEXT NOT NULL,
            internal_signature TEXT,
            customer_id UUID NOT NULL,
            delivery_service TEXT NOT NULL,
            shardkey TEXT NOT NULL,
            sm_id INT NOT NULL,
            date_created TIMESTAMP NOT NULL DEFAULT NOW(),
            oof_shard TEXT
        );
        CREATE INDEX idx_orders_customer_id ON ecommerce.orders(customer_id);
        CREATE INDEX idx_orders_date_created ON ecommerce.orders(date_created);
    END IF;

    -- Таблица order_items
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_tables WHERE schemaname = 'ecommerce' AND tablename = 'order_items') THEN
        CREATE TABLE ecommerce.order_items (
            order_uid UUID REFERENCES ecommerce.orders(order_uid) ON DELETE CASCADE,
            item_id INT REFERENCES ecommerce.items(id) ON DELETE CASCADE,
            PRIMARY KEY (order_uid, item_id)
        );
    END IF;
END$$;