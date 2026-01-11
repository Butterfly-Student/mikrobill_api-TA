-- +goose Up
-- +goose StatementBegin

CREATE TYPE customer_status AS ENUM ('active', 'suspended', 'inactive', 'pending');
CREATE TYPE service_type AS ENUM ('pppoe', 'hotspot', 'static_ip');

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    mikrotik_id UUID NOT NULL,
    package_id UUID,

    -- Identitas pelanggan
    username VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    address VARCHAR(255),

    -- Jenis layanan
    service_type service_type NOT NULL,

    -- Sinkronisasi Mikrotik (WAJIB)
    mikrotik_object_id VARCHAR(50), -- PPP Secret ID / Hotspot User ID

    -- Network info
    assigned_ip INET,
    mac_address MACADDR,
    interface VARCHAR(50),
    last_online TIMESTAMPTZ,
    last_ip INET,

    -- Status & billing
    status customer_status NOT NULL DEFAULT 'inactive',
    auto_suspension BOOLEAN NOT NULL DEFAULT true,
    billing_day INTEGER NOT NULL DEFAULT 15 CHECK (billing_day BETWEEN 1 AND 28),
    join_date TIMESTAMPTZ NOT NULL DEFAULT now(),

    customer_notes VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_customers_mikrotik
        FOREIGN KEY (mikrotik_id)
        REFERENCES mikrotik(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_customers_package
        FOREIGN KEY (package_id)
        REFERENCES packages(id)
        ON DELETE SET NULL,

    CONSTRAINT uq_customer_username_per_mikrotik
        UNIQUE (mikrotik_id, username),

    CONSTRAINT uq_customer_phone_per_mikrotik
        UNIQUE (mikrotik_id, phone),

    CONSTRAINT uq_customer_mikrotik_object
        UNIQUE (mikrotik_id, mikrotik_object_id)
);

CREATE INDEX idx_customers_mikrotik ON customers(mikrotik_id);
CREATE INDEX idx_customers_package ON customers(package_id);
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_service_type ON customers(service_type);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_mac ON customers(mac_address);

CREATE TRIGGER set_updated_at_customers
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd
CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    mikrotik_id UUID NOT NULL,

    name VARCHAR(100) NOT NULL,
    speed VARCHAR(50) NOT NULL,

    price NUMERIC(12,2) NOT NULL CHECK (price >= 0),

    is_active BOOLEAN NOT NULL DEFAULT true,

    -- reference ke PPP Profile di Mikrotik
    ppp_profile_id VARCHAR(50) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_packages_mikrotik
        FOREIGN KEY (mikrotik_id)
        REFERENCES mikrotik(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_package_name_per_mikrotik
        UNIQUE (mikrotik_id, name),

    CONSTRAINT uq_ppp_profile_per_mikrotik
        UNIQUE (mikrotik_id, ppp_profile_id)
);
