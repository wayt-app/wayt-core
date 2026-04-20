-- +migrate Up
ALTER TYPE tabl_booking_status ADD VALUE IF NOT EXISTS 'no_show';

-- +migrate Down
-- PostgreSQL does not support removing enum values.
