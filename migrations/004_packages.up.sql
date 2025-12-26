-- +goose Up
-- +goose StatementBegin

CREATE TYPE profile_type AS ENUM ('pppoe', 'hotspot');
CREATE TYPE service_type AS ENUM ('pppoe', 'hotspot', 'static_ip');


-- Profile dasar (common fields)
CREATE TABLE mikrotik_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    profile_type profile_type NOT NULL,
    
    -- Rate limiting (common untuk semua)
    rate_limit_up VARCHAR(50),
    rate_limit_down VARCHAR(50),
    
    -- Common settings
    idle_timeout VARCHAR(20),
    session_timeout VARCHAR(20),
    keepalive_timeout VARCHAR(20),
    only_one BOOLEAN DEFAULT false,
    status_authentication BOOLEAN DEFAULT true,
    dns_server VARCHAR(100),
    
    is_active BOOLEAN DEFAULT true,
    sync_with_mikrotik BOOLEAN DEFAULT true,
    last_sync TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    UNIQUE (mikrotik_id, name, profile_type)
);

CREATE INDEX idx_profiles_mikrotik ON mikrotik_profiles(mikrotik_id);
CREATE INDEX idx_profiles_type ON mikrotik_profiles(profile_type);
CREATE INDEX idx_profiles_active ON mikrotik_profiles(is_active);

-- Tabel khusus PPPoE
CREATE TABLE mikrotik_profile_pppoe (
    profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
    local_address VARCHAR(50) NOT NULL,
    remote_address VARCHAR(50),
    address_pool VARCHAR(50) NOT NULL,
    mtu VARCHAR(10) DEFAULT '1480',
    mru VARCHAR(10) DEFAULT '1480',
    
    -- PPPoE specific settings
    service_name VARCHAR(50),
    max_mtu VARCHAR(10),
    max_mru VARCHAR(10),
    use_mpls BOOLEAN DEFAULT false,
    use_compression BOOLEAN DEFAULT false,
    use_encryption BOOLEAN DEFAULT false
);

CREATE INDEX idx_profile_pppoe_pool ON mikrotik_profile_pppoe(address_pool);

-- Tabel khusus Hotspot
CREATE TABLE mikrotik_profile_hotspot (
    profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
    shared_users INTEGER DEFAULT 1,
    hotspot_address_pool VARCHAR(50),
    
    -- Hotspot specific
    transparent_proxy BOOLEAN DEFAULT false,
    smtp_server VARCHAR(100),
    http_proxy VARCHAR(100),
    http_cookie_lifetime VARCHAR(20),
    
    -- Authentication
    mac_auth BOOLEAN DEFAULT false,
    mac_auth_mode VARCHAR(20) DEFAULT 'none', -- none, mac-only, mac-and-password
    trial_user_profile VARCHAR(50),
    
    -- Limits
    mac_cookie_timeout VARCHAR(20),
    login_timeout VARCHAR(20)
);

CREATE INDEX idx_profile_hotspot_pool ON mikrotik_profile_hotspot(hotspot_address_pool);

-- Tabel khusus Static IP
CREATE TABLE mikrotik_profile_static_ip (
    profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
    ip_pool VARCHAR(50), -- contoh: '10.10.10.0/24'
    gateway VARCHAR(50) NOT NULL,
    netmask VARCHAR(50) NOT NULL DEFAULT '255.255.255.0',
    
    -- Security
    allowed_mac_addresses TEXT[], -- Array MAC addresses yang diizinkan
    firewall_chain VARCHAR(50),
    
    -- VLAN support
    vlan_id INTEGER,
    vlan_priority INTEGER,
    
    -- Advanced routing
    route_distance INTEGER DEFAULT 1,
    routing_mark VARCHAR(50)
);

CREATE INDEX idx_profile_static_ip_pool ON mikrotik_profile_static_ip(ip_pool);
CREATE INDEX idx_profile_static_ip_vlan ON mikrotik_profile_static_ip(vlan_id);


-- Table khusu juga untuk ip_static saya bingung ingin menggunakan yang mana

CREATE TABLE static_ip_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,

    ip_address INET NOT NULL,
    gateway INET,
    netmask INTEGER,

    allowed_mac_addresses TEXT[],
    firewall_chain VARCHAR(50),
    rate_limit VARCHAR(50),

    is_active BOOLEAN DEFAULT true,
    sync_with_mikrotik BOOLEAN DEFAULT true,
    last_sync TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE (mikrotik_id, ip_address)
);

CREATE INDEX idx_static_ip_assignments_mikrotik ON static_ip_assignments(mikrotik_id);
CREATE INDEX idx_static_ip_assignments_active ON static_ip_assignments(is_active);
CREATE INDEX idx_ip_address_ ON static_ip_assignments(ip_address);

-- Queue settings (shared untuk semua tipe)
CREATE TABLE mikrotik_queue_settings (
    profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
    
    -- Queue type
    queue_type VARCHAR(50) DEFAULT 'default',
    parent_queue VARCHAR(50),
    priority VARCHAR(20) DEFAULT '8',
    
    -- Burst settings
    burst_limit_up VARCHAR(50),
    burst_limit_down VARCHAR(50),
    burst_threshold_up VARCHAR(50),
    burst_threshold_down VARCHAR(50),
    burst_time VARCHAR(20) DEFAULT '0s',
    
    -- Advanced
    limit_at_up VARCHAR(50),
    limit_at_down VARCHAR(50),
    max_limit_up VARCHAR(50),
    max_limit_down VARCHAR(50),
    
    -- Packet marking
    packet_marks TEXT[]
);

-- Packages (billing plans)
CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
    
    -- Package info
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(150),
    description TEXT,
    
    -- Speed & pricing
    speed VARCHAR(50) NOT NULL, -- Display purpose: "10 Mbps"
    bandwidth_up VARCHAR(50), -- Technical: "10M"
    bandwidth_down VARCHAR(50), -- Technical: "10M"
    price DECIMAL(10,2) NOT NULL,
    tax_rate DECIMAL(5,2) DEFAULT 11.00,
    
    -- Service type dan profile
    service_type service_type NOT NULL,
    profile_id UUID REFERENCES mikrotik_profiles(id) ON DELETE SET NULL,
    
    -- Untuk static IP customers (jika tidak pakai profile)
    static_ip_pool VARCHAR(50),
    
    -- Package features
    fup_enabled BOOLEAN DEFAULT false,
    fup_quota INTEGER, -- in GB
    fup_speed VARCHAR(50), -- speed after FUP
    
    -- Validity (untuk prepaid/voucher)
    validity_days INTEGER, -- NULL = monthly subscription
    validity_hours INTEGER, -- untuk time-based vouchers
    
    -- Display & marketing
    image_filename VARCHAR(255),
    is_featured BOOLEAN DEFAULT false,
    is_popular BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    
    -- Package category/tags
    category VARCHAR(50), -- 'home', 'business', 'gaming', etc
    tags TEXT[], -- ['unlimited', 'gaming', 'streaming']
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_packages_mikrotik ON packages(mikrotik_id);
CREATE INDEX idx_packages_service_type ON packages(service_type);
CREATE INDEX idx_packages_is_active ON packages(is_active);
CREATE INDEX idx_packages_is_featured ON packages(is_featured);
CREATE INDEX idx_packages_profile ON packages(profile_id);
CREATE INDEX idx_packages_category ON packages(category);

-- Triggers
CREATE TRIGGER set_updated_at_mikrotik_profiles
    BEFORE UPDATE ON mikrotik_profiles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_packages
    BEFORE UPDATE ON packages
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- Function untuk validasi profile_id sesuai dengan service_type
CREATE OR REPLACE FUNCTION validate_package_profile()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.profile_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM mikrotik_profiles 
            WHERE id = NEW.profile_id 
            AND profile_type = NEW.service_type
        ) THEN
            RAISE EXCEPTION 'Profile type (%) does not match service type (%)', 
                (SELECT profile_type FROM mikrotik_profiles WHERE id = NEW.profile_id),
                NEW.service_type;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger untuk validasi
CREATE TRIGGER validate_package_profile_trigger
    BEFORE INSERT OR UPDATE ON packages
    FOR EACH ROW EXECUTE FUNCTION validate_package_profile();