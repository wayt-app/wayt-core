-- 028_add_token_version.sql
-- Adds token_version to support JWT invalidation on logout.
-- Default 0 so existing tokens (which carry no token_version claim) continue to work
-- until the user explicitly logs out, at which point the version is incremented.

ALTER TABLE tabl_customers        ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 0;
ALTER TABLE tabl_business_owners  ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 0;
ALTER TABLE tabl_staff            ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 0;
ALTER TABLE tabl_admin_users      ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 0;
