CREATE TABLE IF NOT EXISTS tabl_plan_orders (
    id               BIGSERIAL PRIMARY KEY,
    business_owner_id BIGINT NOT NULL REFERENCES tabl_business_owners(id),
    plan_id          BIGINT NOT NULL REFERENCES tabl_plans(id),
    status           VARCHAR(20) NOT NULL DEFAULT 'pending',
    process_at       TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tabl_plan_orders_owner ON tabl_plan_orders(business_owner_id);
CREATE INDEX IF NOT EXISTS idx_tabl_plan_orders_status ON tabl_plan_orders(status, process_at);
