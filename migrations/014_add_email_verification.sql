-- Migration 014: add email verification columns to tabl_customers

ALTER TABLE tabl_customers
    ADD COLUMN IF NOT EXISTS is_verified          BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS verification_token   VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_tabl_customers_verification_token
    ON tabl_customers (verification_token) WHERE verification_token IS NOT NULL;
