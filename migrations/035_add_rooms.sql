-- Ruangan (rooms) untuk mengelompokkan meja dalam satu cabang
CREATE TABLE IF NOT EXISTS tabl_rooms (
    id          BIGSERIAL PRIMARY KEY,
    branch_id   BIGINT        NOT NULL REFERENCES tabl_branches(id),
    name        VARCHAR(100)  NOT NULL,
    is_smoking  BOOLEAN       NOT NULL DEFAULT FALSE,
    is_default  BOOLEAN       NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tabl_rooms_branch_id ON tabl_rooms(branch_id);

-- room_id nullable: NULL = ruangan default (single-room setup)
ALTER TABLE tabl_table_types ADD COLUMN IF NOT EXISTS room_id BIGINT REFERENCES tabl_rooms(id);
ALTER TABLE tabl_bookings    ADD COLUMN IF NOT EXISTS room_id BIGINT REFERENCES tabl_rooms(id);
