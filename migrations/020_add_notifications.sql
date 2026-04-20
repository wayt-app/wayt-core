CREATE TABLE tabl_notifications (
    id SERIAL PRIMARY KEY,
    user_type VARCHAR(20) NOT NULL,  -- customer, owner, staff
    user_id   INT NOT NULL,
    title     VARCHAR(200) NOT NULL,
    message   TEXT NOT NULL,
    is_read   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tabl_notif_user ON tabl_notifications(user_type, user_id, is_read);
