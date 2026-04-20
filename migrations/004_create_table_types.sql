-- +migrate Up
CREATE TABLE IF NOT EXISTS tabl_table_types (
    id           BIGSERIAL    PRIMARY KEY,
    branch_id    BIGINT       NOT NULL REFERENCES tabl_branches(id),
    name         VARCHAR(100) NOT NULL,   -- e.g. "Meja Keluarga", "VIP Booth"
    capacity     INT          NOT NULL,   -- jumlah kursi per meja
    total_tables INT          NOT NULL DEFAULT 1, -- jumlah meja tipe ini
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ  NULL
);

-- +migrate Down
DROP TABLE IF EXISTS table_types;
