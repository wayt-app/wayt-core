CREATE TABLE tabl_media (
    id SERIAL PRIMARY KEY,
    restaurant_id INT NOT NULL REFERENCES tabl_restaurants(id),
    branch_id INT REFERENCES tabl_branches(id), -- NULL = berlaku untuk semua cabang
    type VARCHAR(20) NOT NULL,                   -- 'logo' atau 'menu'
    url TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tabl_media_restaurant ON tabl_media(restaurant_id, type);
CREATE INDEX idx_tabl_media_branch ON tabl_media(branch_id);
