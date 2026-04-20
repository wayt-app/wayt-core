ALTER TABLE tabl_customers ADD COLUMN IF NOT EXISTS google_id VARCHAR(255);
ALTER TABLE tabl_customers ADD COLUMN IF NOT EXISTS avatar_url TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tabl_customers_google_id ON tabl_customers (google_id) WHERE google_id IS NOT NULL;
