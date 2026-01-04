-- -- +goose Up
-- -- +goose StatementBegin

-- -- Hotspot Vouchers Table
-- CREATE TABLE hotspot_vouchers (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
    
--     -- Voucher details
--     code VARCHAR(50) NOT NULL,
--     username VARCHAR(100) NOT NULL,
--     password VARCHAR(100) NOT NULL,
    
--     -- Package & Profile
--     package_id UUID REFERENCES packages(id) ON DELETE SET NULL,
--     profile_id UUID REFERENCES mikrotik_profiles(id) ON DELETE SET NULL,
    
--     -- Pricing
--     price DECIMAL(10,2) NOT NULL,
--     agent_price DECIMAL(10,2), -- harga untuk agent
--     commission DECIMAL(10,2), -- komisi dalam persen
    
--     -- Validity
--     validity_days INTEGER,
--     validity_hours INTEGER,
--     valid_from TIMESTAMPTZ,
--     valid_until TIMESTAMPTZ,
    
--     -- Usage tracking
--     status voucher_status DEFAULT 'active',
--     activated_at TIMESTAMPTZ,
--     activated_by VARCHAR(100), -- customer phone/email
--     expired_at TIMESTAMPTZ,
    
--     -- Who created/sold this voucher
--     created_by UUID REFERENCES users(id) ON DELETE SET NULL,
--     agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    
--     -- MikroTik sync
--     sync_with_mikrotik BOOLEAN DEFAULT true,
--     last_sync TIMESTAMPTZ,
--     mikrotik_user_id VARCHAR(100), -- ID dari MikroTik
    
--     -- Batch info (for bulk generation)
--     batch_id UUID,
--     batch_name VARCHAR(100),
    
--     notes TEXT,
--     created_at TIMESTAMPTZ DEFAULT now(),
--     updated_at TIMESTAMPTZ DEFAULT now(),
    
--     UNIQUE (mikrotik_id, code),
--     UNIQUE (mikrotik_id, username)
-- );

-- CREATE INDEX idx_vouchers_mikrotik ON hotspot_vouchers(mikrotik_id);
-- CREATE INDEX idx_vouchers_code ON hotspot_vouchers(code);
-- CREATE INDEX idx_vouchers_username ON hotspot_vouchers(username);
-- CREATE INDEX idx_vouchers_status ON hotspot_vouchers(status);
-- CREATE INDEX idx_vouchers_package ON hotspot_vouchers(package_id);
-- CREATE INDEX idx_vouchers_agent ON hotspot_vouchers(agent_id);
-- CREATE INDEX idx_vouchers_batch ON hotspot_vouchers(batch_id);
-- CREATE INDEX idx_vouchers_valid_until ON hotspot_vouchers(valid_until);

-- -- Voucher Batches Table (for tracking bulk generation)
-- CREATE TABLE voucher_batches (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
    
--     batch_name VARCHAR(100) NOT NULL,
--     package_id UUID REFERENCES packages(id) ON DELETE SET NULL,
--     profile_id UUID REFERENCES mikrotik_profiles(id) ON DELETE SET NULL,
    
--     total_vouchers INTEGER NOT NULL,
--     price_per_voucher DECIMAL(10,2) NOT NULL,
    
--     -- Prefix/suffix for code generation
--     code_prefix VARCHAR(20),
--     code_suffix VARCHAR(20),
--     code_length INTEGER DEFAULT 8,
    
--     created_by UUID REFERENCES users(id) ON DELETE SET NULL,
--     created_at TIMESTAMPTZ DEFAULT now()
-- );

-- CREATE INDEX idx_batches_mikrotik ON voucher_batches(mikrotik_id);
-- CREATE INDEX idx_batches_package ON voucher_batches(package_id);

-- -- Voucher Usage Logs
-- CREATE TABLE voucher_usage_logs (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     voucher_id UUID NOT NULL REFERENCES hotspot_vouchers(id) ON DELETE CASCADE,
    
--     event_type VARCHAR(50) NOT NULL, -- activated, logged_in, logged_out, expired
--     ip_address INET,
--     mac_address MACADDR,
--     session_id VARCHAR(100),
    
--     upload_bytes BIGINT DEFAULT 0,
--     download_bytes BIGINT DEFAULT 0,
--     total_bytes BIGINT DEFAULT 0,
    
--     session_time INTEGER, -- in seconds
--     event_time TIMESTAMPTZ DEFAULT now()
-- );

-- CREATE INDEX idx_voucher_logs_voucher ON voucher_usage_logs(voucher_id);
-- CREATE INDEX idx_voucher_logs_event ON voucher_usage_logs(event_type);
-- CREATE INDEX idx_voucher_logs_time ON voucher_usage_logs(event_time);

-- CREATE TRIGGER set_updated_at_hotspot_vouchers
--     BEFORE UPDATE ON hotspot_vouchers
--     FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -- +goose StatementEnd