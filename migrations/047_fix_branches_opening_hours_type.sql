-- opening_hours was created as JSONB in migration 003 but the app writes plain text.
-- Convert to TEXT to match the model definition.
ALTER TABLE tabl_branches
    ALTER COLUMN opening_hours TYPE TEXT USING opening_hours::text;
