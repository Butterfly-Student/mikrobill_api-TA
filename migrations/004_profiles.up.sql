-- +goose Up
-- +goose StatementBegin

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
    
    -- Custom billing field
    price DECIMAL(15,2),
    
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

CREATE TRIGGER set_updated_at_mikrotik_profiles
    BEFORE UPDATE ON mikrotik_profiles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd