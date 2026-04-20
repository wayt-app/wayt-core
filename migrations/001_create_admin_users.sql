-- +migrate Up
CREATE TYPE tabl_admin_role AS ENUM ('superadmin', 'admin');

CREATE TABLE IF NOT EXISTS tabl_admin_users (
    id            BIGSERIAL    PRIMARY KEY,
    username      VARCHAR(100) NOT NULL UNIQUE,
    password      VARCHAR(255) NOT NULL,
    role          tabl_admin_role   NOT NULL DEFAULT 'admin',
    restaurant_id BIGINT       NULL,    -- filled after restaurants table exists
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS admin_users;
DROP TYPE IF EXISTS tabl_admin_role;
