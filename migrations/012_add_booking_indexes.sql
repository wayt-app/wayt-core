-- +migrate Up

-- Fix slow query: FindNoShowCandidates runs every minute, scans by status + booking_date.
-- status first (very selective for 'confirmed'), then booking_date + start_time.
CREATE INDEX IF NOT EXISTS idx_tabl_bookings_noshow
    ON tabl_bookings (status, booking_date, start_time);

-- Cover branch-scoped queries: FindByBranch, FindWaitingListForSlot,
-- CountByStatusForDate, CountOverlappingByTableType.
CREATE INDEX IF NOT EXISTS idx_tabl_bookings_branch_date
    ON tabl_bookings (branch_id, booking_date, status);

-- Cover FindByCustomer.
CREATE INDEX IF NOT EXISTS idx_tabl_bookings_customer
    ON tabl_bookings (customer_id);

-- tabl_branches: FindByRestaurant / FindActiveByRestaurant filter by restaurant_id.
CREATE INDEX IF NOT EXISTS idx_tabl_branches_restaurant
    ON tabl_branches (restaurant_id);

-- tabl_table_types: FindByBranch filters by branch_id (called on every slot/availability check).
CREATE INDEX IF NOT EXISTS idx_tabl_table_types_branch
    ON tabl_table_types (branch_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_tabl_bookings_noshow;
DROP INDEX IF EXISTS idx_tabl_bookings_branch_date;
DROP INDEX IF EXISTS idx_tabl_bookings_customer;
DROP INDEX IF EXISTS idx_tabl_branches_restaurant;
DROP INDEX IF EXISTS idx_tabl_table_types_branch;
