-- +migrate Up

-- tabl_customers: phone lookup (FindByPhone, WhatsApp OTP)
CREATE INDEX IF NOT EXISTS idx_tabl_customers_phone
    ON tabl_customers (phone);

-- tabl_email_campaigns: FindDue cron filters WHERE status='scheduled' AND scheduled_at <= now()
-- Composite lebih efisien dari single-column status saja
CREATE INDEX IF NOT EXISTS idx_tabl_email_campaigns_due
    ON tabl_email_campaigns (status, scheduled_at)
    WHERE scheduled_at IS NOT NULL;

-- tabl_notifications: semua query filter user_type + user_id + is_read secara bersamaan
-- Ganti dua index terpisah (user_type, user_id) menjadi satu composite dengan is_read
CREATE INDEX IF NOT EXISTS idx_tabl_notifications_user_unread
    ON tabl_notifications (user_type, user_id, is_read);

-- pg_trgm extension untuk ILIKE search (booking search by nama/telepon customer)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- GIN index pada tabl_customers untuk ILIKE '%...%' search di FindByBranchPaged
CREATE INDEX IF NOT EXISTS idx_tabl_customers_name_trgm
    ON tabl_customers USING GIN (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_tabl_customers_phone_trgm
    ON tabl_customers USING GIN (phone gin_trgm_ops);

-- +migrate Down
DROP INDEX IF EXISTS idx_tabl_customers_phone;
DROP INDEX IF EXISTS idx_tabl_email_campaigns_due;
DROP INDEX IF EXISTS idx_tabl_notifications_user_unread;
DROP INDEX IF EXISTS idx_tabl_customers_name_trgm;
DROP INDEX IF EXISTS idx_tabl_customers_phone_trgm;
