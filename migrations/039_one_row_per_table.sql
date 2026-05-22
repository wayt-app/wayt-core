-- Expand existing multi-table types into individual rows (1 row = 1 physical table)
DO $$
DECLARE
  r RECORD;
  n INTEGER;
  i INTEGER;
BEGIN
  FOR r IN
    SELECT id, branch_id, room_id, name, capacity, is_active
    FROM tabl_table_types
    WHERE deleted_at IS NULL AND total_tables > 1
  LOOP
    SELECT total_tables INTO n FROM tabl_table_types WHERE id = r.id;
    FOR i IN 2..n LOOP
      INSERT INTO tabl_table_types (branch_id, room_id, name, capacity, is_active, created_at, updated_at)
      VALUES (r.branch_id, r.room_id, r.name, r.capacity, r.is_active, NOW(), NOW());
    END LOOP;
  END LOOP;
END $$;

ALTER TABLE tabl_table_types DROP COLUMN IF EXISTS total_tables;
