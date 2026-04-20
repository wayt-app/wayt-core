-- +migrate Up
CREATE TABLE IF NOT EXISTS tabl_branches (
    id            BIGSERIAL    PRIMARY KEY,
    restaurant_id BIGINT       NOT NULL REFERENCES tabl_restaurants(id),
    name          VARCHAR(150) NOT NULL,
    address       TEXT         NULL,
    phone         VARCHAR(20)  NULL,
    opening_hours JSONB        NULL,   -- e.g. {"mon":"09:00-22:00","tue":"09:00-22:00"}
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL
);

-- +migrate Down
DROP TABLE IF EXISTS branches;
