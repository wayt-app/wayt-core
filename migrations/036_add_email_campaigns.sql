-- Kampanye email blast untuk owner
CREATE TABLE IF NOT EXISTS tabl_email_campaigns (
    id               BIGSERIAL    PRIMARY KEY,
    restaurant_id    BIGINT       NOT NULL REFERENCES tabl_restaurants(id),
    subject          VARCHAR(200) NOT NULL,
    body             TEXT         NOT NULL,
    filter_branch_id BIGINT       REFERENCES tabl_branches(id),
    filter_segment   VARCHAR(20)  NOT NULL DEFAULT 'all',
    status           VARCHAR(20)  NOT NULL DEFAULT 'scheduled',
    scheduled_at     TIMESTAMPTZ,
    sent_at          TIMESTAMPTZ,
    recipient_count  INT          NOT NULL DEFAULT 0,
    success_count    INT          NOT NULL DEFAULT 0,
    fail_count       INT          NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tabl_email_campaigns_restaurant_id ON tabl_email_campaigns(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_tabl_email_campaigns_status        ON tabl_email_campaigns(status);

-- Quota kampanye di plan dan subscription
ALTER TABLE tabl_plans        ADD COLUMN IF NOT EXISTS max_campaigns_per_month INT NOT NULL DEFAULT 0;
ALTER TABLE tabl_subscriptions ADD COLUMN IF NOT EXISTS campaigns_this_month   INT NOT NULL DEFAULT 0;
