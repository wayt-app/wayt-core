CREATE TABLE IF NOT EXISTS tabl_menu_items (
    id                    BIGSERIAL PRIMARY KEY,
    branch_id             BIGINT      NOT NULL REFERENCES tabl_branches(id),
    name                  VARCHAR(100) NOT NULL,
    description           TEXT        NOT NULL DEFAULT '',
    price                 BIGINT      NOT NULL DEFAULT 0,
    image_url             TEXT        NOT NULL DEFAULT '',
    category              VARCHAR(20) NOT NULL,
    is_available          BOOLEAN     NOT NULL DEFAULT TRUE,
    is_favorite           BOOLEAN     NOT NULL DEFAULT FALSE,
    is_chef_recommendation BOOLEAN    NOT NULL DEFAULT FALSE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tabl_menu_items_branch_id  ON tabl_menu_items(branch_id);
CREATE INDEX IF NOT EXISTS idx_tabl_menu_items_deleted_at ON tabl_menu_items(deleted_at);
