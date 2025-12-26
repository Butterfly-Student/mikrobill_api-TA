-- +goose Up
-- +goose StatementBegin

CREATE TYPE customer_status AS ENUM ('active', 'suspended', 'inactive', 'pending');
CREATE TYPE cable_status AS ENUM ('connected', 'disconnected', 'maintenance', 'damaged');
CREATE TYPE service_type AS ENUM ('pppoe', 'hotspot', 'static_ip');

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
    
    -- Basic info
    username VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    address TEXT,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    
    -- Service configuration
    service_type service_type NOT NULL, -- pppoe, hotspot, static_ip
    package_id UUID REFERENCES packages(id) ON DELETE SET NULL,
    
    -- PPPoE specific
    pppoe_username VARCHAR(100),
    pppoe_password VARCHAR(100),
    pppoe_profile_id UUID REFERENCES mikrotik_profiles(id) ON DELETE SET NULL,
    
    -- Hotspot specific
    hotspot_username VARCHAR(100),
    hotspot_password VARCHAR(100),
    hotspot_profile_id UUID REFERENCES mikrotik_profiles(id) ON DELETE SET NULL,
    hotspot_mac_address MACADDR,
    hotspot_ip_address INET,
    
    -- Static IP specific
    static_ip INET,
    static_ip_netmask VARCHAR(20),
    static_ip_gateway INET,
    static_ip_dns1 INET,
    static_ip_dns2 INET,
    
    -- Network info
    assigned_ip INET,
    mac_address MACADDR,
    last_online TIMESTAMPTZ,
    last_ip INET,
    
    -- Connection to infrastructure
    odp_id UUID REFERENCES odps(id) ON DELETE SET NULL,
    
    -- Status & billing
    status customer_status DEFAULT 'active',
    auto_suspension BOOLEAN DEFAULT true,
    billing_day INTEGER DEFAULT 15,
    join_date TIMESTAMPTZ DEFAULT now(),
    
    -- Cable connection fields
    cable_type VARCHAR(50),
    cable_length INTEGER,
    port_number INTEGER,
    cable_status cable_status DEFAULT 'connected',
    cable_notes TEXT,
    
    -- FUP tracking for packages with FUP
    fup_quota_used INTEGER DEFAULT 0, -- in MB
    fup_reset_date DATE,
    is_fup_active BOOLEAN DEFAULT false,
    
    customer_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE (mikrotik_id, username),
    UNIQUE (mikrotik_id, phone),
    UNIQUE (mikrotik_id, pppoe_username),
    UNIQUE (mikrotik_id, hotspot_username),
    UNIQUE (mikrotik_id, static_ip)
);

CREATE INDEX idx_customers_mikrotik ON customers(mikrotik_id);
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_pppoe ON customers(pppoe_username);
CREATE INDEX idx_customers_hotspot ON customers(hotspot_username);
CREATE INDEX idx_customers_static_ip ON customers(static_ip);
CREATE INDEX idx_customers_service_type ON customers(service_type);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_package ON customers(package_id);
CREATE INDEX idx_customers_odp ON customers(odp_id);
CREATE INDEX idx_customers_mac ON customers(mac_address);

CREATE TRIGGER set_updated_at_customers
    BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

