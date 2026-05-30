-- +migrate Up

-- Business owner token lookups (password reset, email verification)
CREATE INDEX IF NOT EXISTS idx_tabl_business_owners_reset_token
    ON tabl_business_owners (reset_token) WHERE reset_token IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tabl_business_owners_verification_token
    ON tabl_business_owners (verification_token) WHERE verification_token IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_tabl_business_owners_reset_token;
DROP INDEX IF EXISTS idx_tabl_business_owners_verification_token;
