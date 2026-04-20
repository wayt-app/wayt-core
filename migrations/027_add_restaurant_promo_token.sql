-- Add promo_token to restaurants for obfuscated public promo URLs
ALTER TABLE tabl_restaurants
    ADD COLUMN IF NOT EXISTS promo_token VARCHAR(32);

-- Generate tokens for existing rows
UPDATE tabl_restaurants
SET promo_token = md5(random()::text || id::text)
WHERE promo_token IS NULL OR promo_token = '';

-- Enforce uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS idx_tabl_restaurants_promo_token
    ON tabl_restaurants (promo_token);
