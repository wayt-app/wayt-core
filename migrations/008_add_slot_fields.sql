-- +migrate Up
ALTER TABLE tabl_branches
    ADD COLUMN open_from            VARCHAR(5) NULL,      -- "HH:MM", e.g. "10:00"
    ADD COLUMN open_to              VARCHAR(5) NULL,      -- "HH:MM", e.g. "22:00"
    ADD COLUMN slot_interval_minutes INT NOT NULL DEFAULT 30;

-- +migrate Down
ALTER TABLE tabl_branches
    DROP COLUMN IF EXISTS open_from,
    DROP COLUMN IF EXISTS open_to,
    DROP COLUMN IF EXISTS slot_interval_minutes;
