-- +migrate Up
ALTER TABLE tabl_branches
    ADD COLUMN default_duration_minutes INT NOT NULL DEFAULT 120,
    ADD COLUMN require_confirmation     BOOLEAN NOT NULL DEFAULT TRUE;

-- +migrate Down
ALTER TABLE tabl_branches
    DROP COLUMN IF EXISTS default_duration_minutes,
    DROP COLUMN IF EXISTS require_confirmation;
