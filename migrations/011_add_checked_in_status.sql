-- +migrate Up
ALTER TYPE tabl_booking_status ADD VALUE IF NOT EXISTS 'checked_in';

-- +migrate Down
-- PostgreSQL does not support removing enum values.
