-- +migrate Up
ALTER TYPE tabl_booking_status ADD VALUE IF NOT EXISTS 'waiting_list';

-- +migrate Down
-- PostgreSQL does not support removing enum values; handled by full schema reset if needed.
