ALTER TABLE tabl_business_owners
  ADD COLUMN IF NOT EXISTS google_id VARCHAR(255),
  ADD COLUMN IF NOT EXISTS avatar_url TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tabl_business_owners_google_id
  ON tabl_business_owners(google_id)
  WHERE google_id IS NOT NULL;
