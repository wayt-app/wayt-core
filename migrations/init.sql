-- =============================================================
-- Tabl — Full Database Init
-- Jalankan sekali di database kosong:
--   psql -U <user> -d tabl -f migrations/init.sql
-- =============================================================

-- Admin users
CREATE TYPE tabl_admin_role AS ENUM ('superadmin', 'admin');

CREATE TABLE IF NOT EXISTS tabl_admin_users (
    id                      BIGSERIAL       PRIMARY KEY,
    username                VARCHAR(100)    NOT NULL UNIQUE,
    password                VARCHAR(255)    NOT NULL,
    role                    tabl_admin_role NOT NULL DEFAULT 'admin',
    restaurant_id           BIGINT          NULL,
    reset_token             VARCHAR(64)     NULL,
    reset_token_expires_at  TIMESTAMPTZ     NULL,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Restaurants
CREATE TABLE IF NOT EXISTS tabl_restaurants (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(150) NOT NULL,
    description TEXT         NULL,
    address     TEXT         NULL,
    phone       VARCHAR(20)  NULL,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ  NULL
);

ALTER TABLE tabl_admin_users
    ADD CONSTRAINT fk_admin_restaurant
    FOREIGN KEY (restaurant_id) REFERENCES tabl_restaurants(id);

-- Branches
CREATE TABLE IF NOT EXISTS tabl_branches (
    id                      BIGSERIAL    PRIMARY KEY,
    restaurant_id           BIGINT       NOT NULL REFERENCES tabl_restaurants(id),
    name                    VARCHAR(150) NOT NULL,
    address                 TEXT         NULL,
    phone                   VARCHAR(20)  NULL,
    opening_hours           TEXT         NULL,
    default_duration_minutes INT         NOT NULL DEFAULT 120,
    require_confirmation    BOOLEAN      NOT NULL DEFAULT TRUE,
    is_active               BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ  NULL
);

-- Table types
CREATE TABLE IF NOT EXISTS tabl_table_types (
    id           BIGSERIAL    PRIMARY KEY,
    branch_id    BIGINT       NOT NULL REFERENCES tabl_branches(id),
    name         VARCHAR(100) NOT NULL,
    capacity     INT          NOT NULL,
    total_tables INT          NOT NULL DEFAULT 1,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ  NULL
);

-- Customers
CREATE TABLE IF NOT EXISTS tabl_customers (
    id                      BIGSERIAL    PRIMARY KEY,
    name                    VARCHAR(100) NOT NULL,
    email                   VARCHAR(150) NOT NULL UNIQUE,
    phone                   VARCHAR(20)  NOT NULL,
    password                VARCHAR(255) NOT NULL,
    is_verified             BOOLEAN      NOT NULL DEFAULT FALSE,
    verification_token      VARCHAR(64)  NULL,
    reset_token             VARCHAR(64)  NULL,
    reset_token_expires_at  TIMESTAMPTZ  NULL,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Bookings
CREATE TYPE tabl_booking_status AS ENUM ('pending', 'confirmed', 'completed', 'cancelled');

CREATE TABLE IF NOT EXISTS tabl_bookings (
    id            BIGSERIAL           PRIMARY KEY,
    customer_id   BIGINT              NOT NULL REFERENCES tabl_customers(id),
    branch_id     BIGINT              NOT NULL REFERENCES tabl_branches(id),
    table_type_id BIGINT              NOT NULL REFERENCES tabl_table_types(id),
    booking_date  DATE                NOT NULL,
    start_time    TIME                NOT NULL,
    end_time      TIME                NOT NULL,
    guest_count   INT                 NOT NULL,
    status        tabl_booking_status NOT NULL DEFAULT 'pending',
    notes         TEXT                NULL,
    created_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tabl_bookings_availability
    ON tabl_bookings (table_type_id, booking_date, status);
