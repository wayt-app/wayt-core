-- +migrate Up
CREATE TYPE tabl_booking_status AS ENUM ('pending', 'confirmed', 'completed', 'cancelled');

CREATE TABLE IF NOT EXISTS tabl_bookings (
    id            BIGSERIAL       PRIMARY KEY,
    customer_id   BIGINT          NOT NULL REFERENCES tabl_customers(id),
    branch_id     BIGINT          NOT NULL REFERENCES tabl_branches(id),
    table_type_id BIGINT          NOT NULL REFERENCES tabl_table_types(id),
    booking_date  DATE            NOT NULL,
    start_time    TIME            NOT NULL,
    end_time      TIME            NOT NULL,
    guest_count   INT             NOT NULL,
    status        tabl_booking_status  NOT NULL DEFAULT 'pending',
    notes         TEXT            NULL,
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Index for fast availability queries
CREATE INDEX idx_tabl_bookings_availability
    ON tabl_bookings (table_type_id, booking_date, status);

-- +migrate Down
DROP INDEX IF EXISTS idx_bookings_availability;
DROP TABLE IF EXISTS bookings;
DROP TYPE IF EXISTS booking_status;
