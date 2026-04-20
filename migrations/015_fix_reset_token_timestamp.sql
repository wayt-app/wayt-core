-- Migration 015: fix reset_token_expires_at column type dari TIMESTAMP ke TIMESTAMPTZ

ALTER TABLE tabl_customers
    ALTER COLUMN reset_token_expires_at TYPE TIMESTAMPTZ
    USING reset_token_expires_at AT TIME ZONE 'UTC';

ALTER TABLE tabl_admin_users
    ALTER COLUMN reset_token_expires_at TYPE TIMESTAMPTZ
    USING reset_token_expires_at AT TIME ZONE 'UTC';
