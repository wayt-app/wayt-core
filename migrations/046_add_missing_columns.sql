-- tabl_restaurants: banner_url
ALTER TABLE tabl_restaurants ADD COLUMN IF NOT EXISTS banner_url TEXT NOT NULL DEFAULT '';

-- tabl_bookings: kolom yang ditambahkan ke model tanpa migration
ALTER TABLE tabl_bookings
    ADD COLUMN IF NOT EXISTS completion_notes TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS total_bill       BIGINT      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS source           VARCHAR(20) NOT NULL DEFAULT 'app',
    ADD COLUMN IF NOT EXISTS guest_email      VARCHAR(150) NOT NULL DEFAULT '';

-- tabl_reviews: tabel baru untuk ulasan customer
CREATE TABLE IF NOT EXISTS tabl_reviews (
    id            BIGSERIAL PRIMARY KEY,
    customer_id   BIGINT    NOT NULL REFERENCES tabl_customers(id),
    restaurant_id BIGINT    NOT NULL REFERENCES tabl_restaurants(id),
    branch_id     BIGINT    NOT NULL REFERENCES tabl_branches(id),
    booking_id    BIGINT    NOT NULL UNIQUE REFERENCES tabl_bookings(id),
    rating        INT       NOT NULL,
    comment       TEXT      NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tabl_reviews_customer_id   ON tabl_reviews (customer_id);
CREATE INDEX IF NOT EXISTS idx_tabl_reviews_restaurant_id ON tabl_reviews (restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tabl_reviews_branch_id     ON tabl_reviews (branch_id);

-- tabl_email_config: konfigurasi template email (singleton, 1 row)
CREATE TABLE IF NOT EXISTS tabl_email_config (
    id              BIGSERIAL PRIMARY KEY,
    header_image_url TEXT NOT NULL DEFAULT '',
    logo_url         TEXT NOT NULL DEFAULT '',
    instagram_url    TEXT NOT NULL DEFAULT '',
    facebook_url     TEXT NOT NULL DEFAULT '',
    tiktok_url       TEXT NOT NULL DEFAULT '',
    website_url      TEXT NOT NULL DEFAULT 'https://wayt.fun',
    support_email    VARCHAR(150) NOT NULL DEFAULT 'support@wayt.fun',
    footer_bg_url    TEXT NOT NULL DEFAULT '',
    footer_note      TEXT NOT NULL DEFAULT 'Email ini dikirim secara otomatis, mohon tidak membalas email ini.',
    copyright        TEXT NOT NULL DEFAULT '© 2026 Wayt. All rights reserved.',
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed 1 row default jika belum ada
INSERT INTO tabl_email_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
