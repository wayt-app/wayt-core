-- +migrate Up
CREATE TABLE IF NOT EXISTS tabl_restaurants (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(150) NOT NULL,
    description TEXT         NULL,
    address     TEXT         NULL,
    phone       VARCHAR(20)  NULL,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ  NULL
);

-- FK from admin_users.restaurant_id
ALTER TABLE tabl_admin_users
    ADD CONSTRAINT fk_admin_restaurant
    FOREIGN KEY (restaurant_id) REFERENCES tabl_restaurants(id);

-- +migrate Down
ALTER TABLE tabl_admin_users DROP CONSTRAINT IF EXISTS fk_admin_restaurant;
DROP TABLE IF EXISTS restaurants;
