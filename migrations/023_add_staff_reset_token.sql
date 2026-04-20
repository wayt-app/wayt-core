ALTER TABLE tabl_staff ADD COLUMN IF NOT EXISTS reset_token VARCHAR(64);
ALTER TABLE tabl_staff ADD COLUMN IF NOT EXISTS reset_token_expires_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_tabl_staff_reset_token ON tabl_staff(reset_token);
