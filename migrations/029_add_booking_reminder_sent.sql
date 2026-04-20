-- 029_add_booking_reminder_sent.sql
-- Tracks whether the H-1 reminder has been sent for a booking so we
-- never send it twice even if the background job restarts.

ALTER TABLE tabl_bookings ADD COLUMN IF NOT EXISTS reminder_sent BOOLEAN NOT NULL DEFAULT FALSE;
