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
    id               BIGSERIAL    PRIMARY KEY,
    name             VARCHAR(150) NOT NULL,
    description      TEXT         NULL,
    address          TEXT         NULL,
    phone            VARCHAR(20)  NULL,
    cuisine_type     VARCHAR(50)  NOT NULL DEFAULT '',
    logo_url         TEXT         NOT NULL DEFAULT '',
    banner_url       TEXT         NOT NULL DEFAULT '',
    promo_token      VARCHAR(32)  NULL,
    business_owner_id BIGINT      NULL,
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ  NULL
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
    token_version           INT          NOT NULL DEFAULT 0,
    verification_token      VARCHAR(64)  NULL,
    reset_token             VARCHAR(64)  NULL,
    reset_token_expires_at  TIMESTAMPTZ  NULL,
    google_id               VARCHAR(255) NULL,
    avatar_url              TEXT         NULL,
    address                 TEXT         NULL,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tabl_customers_google_id ON tabl_customers (google_id) WHERE google_id IS NOT NULL;

-- Bookings
CREATE TYPE tabl_booking_status AS ENUM ('pending', 'confirmed', 'completed', 'cancelled', 'waiting_list', 'no_show', 'checked_in');

CREATE TABLE IF NOT EXISTS tabl_bookings (
    id                BIGSERIAL           PRIMARY KEY,
    customer_id       BIGINT              NOT NULL REFERENCES tabl_customers(id),
    branch_id         BIGINT              NOT NULL REFERENCES tabl_branches(id),
    table_type_id     BIGINT              NOT NULL REFERENCES tabl_table_types(id),
    room_id           BIGINT              NULL,
    booking_date      DATE                NOT NULL,
    start_time        TIME                NOT NULL,
    end_time          TIME                NOT NULL,
    guest_count       INT                 NOT NULL,
    tables_count      INT                 NOT NULL DEFAULT 1,
    status            tabl_booking_status NOT NULL DEFAULT 'pending',
    notes             TEXT                NULL,
    menu_order        TEXT                NOT NULL DEFAULT '',
    order_status      VARCHAR(20)         NOT NULL DEFAULT 'new',
    payment_proof_url VARCHAR(500)        NOT NULL DEFAULT '',
    cancel_reason     TEXT                NULL,
    completion_notes  TEXT                NOT NULL DEFAULT '',
    total_bill        BIGINT              NOT NULL DEFAULT 0,
    source            VARCHAR(20)         NOT NULL DEFAULT 'app',
    guest_email       VARCHAR(150)        NOT NULL DEFAULT '',
    is_over_limit     BOOLEAN             NOT NULL DEFAULT FALSE,
    reminder_sent     BOOLEAN             NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tabl_bookings_availability ON tabl_bookings (table_type_id, booking_date, status);

-- Reviews
CREATE TABLE IF NOT EXISTS tabl_reviews (
    id            BIGSERIAL PRIMARY KEY,
    customer_id   BIGINT    NOT NULL REFERENCES tabl_customers(id),
    restaurant_id BIGINT    NOT NULL REFERENCES tabl_restaurants(id),
    branch_id     BIGINT    NOT NULL REFERENCES tabl_branches(id),
    booking_id    BIGINT    NOT NULL UNIQUE REFERENCES tabl_bookings(id),
    rating        INT       NOT NULL,
    comment       TEXT      NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tabl_reviews_customer_id   ON tabl_reviews (customer_id);
CREATE INDEX IF NOT EXISTS idx_tabl_reviews_restaurant_id ON tabl_reviews (restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tabl_reviews_branch_id     ON tabl_reviews (branch_id);

-- Email config (singleton)
CREATE TABLE IF NOT EXISTS tabl_email_config (
    id               BIGSERIAL PRIMARY KEY,
    header_image_url TEXT         NOT NULL DEFAULT '',
    logo_url         TEXT         NOT NULL DEFAULT '',
    instagram_url    TEXT         NOT NULL DEFAULT '',
    facebook_url     TEXT         NOT NULL DEFAULT '',
    tiktok_url       TEXT         NOT NULL DEFAULT '',
    website_url      TEXT         NOT NULL DEFAULT 'https://wayt.fun',
    support_email    VARCHAR(150) NOT NULL DEFAULT 'support@wayt.fun',
    footer_bg_url    TEXT         NOT NULL DEFAULT '',
    footer_note      TEXT         NOT NULL DEFAULT 'Email ini dikirim secara otomatis, mohon tidak membalas email ini.',
    copyright        TEXT         NOT NULL DEFAULT '© 2026 Wayt. All rights reserved.',
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO tabl_email_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- Promo links
CREATE TABLE IF NOT EXISTS tabl_promo_links (
    id            BIGSERIAL    PRIMARY KEY,
    restaurant_id BIGINT       NOT NULL REFERENCES tabl_restaurants(id),
    label         VARCHAR(100) NOT NULL DEFAULT '',
    token         VARCHAR(32)  NOT NULL,
    visit_count   BIGINT       NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tabl_promo_links_token         ON tabl_promo_links (token);
CREATE INDEX IF NOT EXISTS idx_tabl_promo_links_restaurant_id        ON tabl_promo_links (restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tabl_promo_links_deleted_at           ON tabl_promo_links (deleted_at);
