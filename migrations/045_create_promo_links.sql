CREATE TABLE IF NOT EXISTS tabl_promo_links (
    id            BIGSERIAL    PRIMARY KEY,
    restaurant_id BIGINT       NOT NULL REFERENCES tabl_restaurants(id),
    label         VARCHAR(100) NOT NULL DEFAULT '',
    token         VARCHAR(32)  NOT NULL,
    visit_count   BIGINT       NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tabl_promo_links_token ON tabl_promo_links (token);
CREATE INDEX IF NOT EXISTS idx_tabl_promo_links_restaurant_id ON tabl_promo_links (restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tabl_promo_links_deleted_at ON tabl_promo_links (deleted_at);
