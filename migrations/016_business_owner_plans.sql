-- Plans
CREATE TABLE tabl_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    max_branches INT NOT NULL DEFAULT 1,
    max_reservations_per_month INT NOT NULL DEFAULT 15,
    wa_notif_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    warning_threshold_pct INT NOT NULL DEFAULT 80,
    price NUMERIC(12,2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO tabl_plans (name, max_branches, max_reservations_per_month, wa_notif_enabled, warning_threshold_pct, price) VALUES
('Starter', 1, 15, false, 80, 99000),
('Growth', 3, 100, true, 80, 299000),
('Pro', 10, 300, true, 80, 799000),
('Enterprise', -1, -1, true, 80, 1999000);

-- Business Owners
CREATE TABLE tabl_business_owners (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    phone VARCHAR(20),
    password VARCHAR(255) NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verification_token VARCHAR(64),
    reset_token VARCHAR(64),
    reset_token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Subscriptions
CREATE TYPE tabl_subscription_status AS ENUM ('trial', 'pending_approval', 'active', 'suspended', 'expired');
CREATE TABLE tabl_subscriptions (
    id SERIAL PRIMARY KEY,
    business_owner_id INT NOT NULL REFERENCES tabl_business_owners(id),
    plan_id INT NOT NULL REFERENCES tabl_plans(id),
    status tabl_subscription_status NOT NULL DEFAULT 'trial',
    trial_started_at TIMESTAMPTZ,
    trial_ends_at TIMESTAMPTZ,
    activated_at TIMESTAMPTZ,
    notes TEXT,
    reservations_this_month INT NOT NULL DEFAULT 0,
    last_reset_at DATE NOT NULL DEFAULT CURRENT_DATE,
    warning_sent BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Staff
CREATE TABLE tabl_staff (
    id SERIAL PRIMARY KEY,
    business_owner_id INT NOT NULL REFERENCES tabl_business_owners(id),
    branch_id INT NOT NULL REFERENCES tabl_branches(id),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Link restaurant to business owner
ALTER TABLE tabl_restaurants ADD COLUMN IF NOT EXISTS business_owner_id INT REFERENCES tabl_business_owners(id);
CREATE INDEX IF NOT EXISTS idx_tabl_restaurants_owner ON tabl_restaurants(business_owner_id);
