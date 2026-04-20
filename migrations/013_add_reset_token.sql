-- Migration 013: add reset_token columns for forgot password

ALTER TABLE tabl_customers
    ADD COLUMN IF NOT EXISTS reset_token VARCHAR(64),
    ADD COLUMN IF NOT EXISTS reset_token_expires_at TIMESTAMP;

ALTER TABLE tabl_admin_users
    ADD COLUMN IF NOT EXISTS reset_token VARCHAR(64),
    ADD COLUMN IF NOT EXISTS reset_token_expires_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_tabl_customers_reset_token
    ON tabl_customers (reset_token) WHERE reset_token IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tabl_admin_users_reset_token
    ON tabl_admin_users (reset_token) WHERE reset_token IS NOT NULL;
