-- +migrate Up

-- Referensi booking terkait untuk tiap notifikasi (nullable — semua notif saat ini
-- terkait booking, tapi dibuat nullable agar notif non-booking tetap bisa dibuat).
ALTER TABLE tabl_notifications ADD COLUMN IF NOT EXISTS booking_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_tabl_notifications_booking_id
    ON tabl_notifications (booking_id) WHERE booking_id IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_tabl_notifications_booking_id;
ALTER TABLE tabl_notifications DROP COLUMN IF EXISTS booking_id;
