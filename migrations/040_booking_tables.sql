CREATE TABLE IF NOT EXISTS tabl_booking_tables (
  id            SERIAL PRIMARY KEY,
  booking_id    INTEGER NOT NULL REFERENCES tabl_bookings(id),
  table_type_id INTEGER NOT NULL REFERENCES tabl_table_types(id),
  count         INTEGER NOT NULL DEFAULT 1,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_booking_tables_booking ON tabl_booking_tables(booking_id);
CREATE INDEX IF NOT EXISTS idx_booking_tables_type    ON tabl_booking_tables(table_type_id);

-- Migrate existing active/pending bookings
INSERT INTO tabl_booking_tables (booking_id, table_type_id, count)
SELECT id, table_type_id, COALESCE(tables_count, 1)
FROM tabl_bookings
WHERE status IN ('pending','confirmed','checked_in','waiting_list');
