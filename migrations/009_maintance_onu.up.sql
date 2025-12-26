-- +goose Up
-- +goose StatementBegin

CREATE TYPE maintenance_type AS ENUM ('repair', 'replacement', 'inspection', 'upgrade');
CREATE TYPE device_status AS ENUM ('online', 'offline', 'maintenance');

-- CABLE MAINTENANCE LOGS
CREATE TABLE cable_maintenance_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cable_route_id UUID REFERENCES cable_routes(id) ON DELETE CASCADE,
    network_segment_id UUID REFERENCES network_segments(id) ON DELETE CASCADE,
    maintenance_type maintenance_type NOT NULL,
    description TEXT NOT NULL,
    performed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    maintenance_date DATE NOT NULL,
    duration_hours DECIMAL(4,2),
    cost DECIMAL(12,2),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_maintenance_logs_date ON cable_maintenance_logs(maintenance_date);
CREATE INDEX idx_maintenance_logs_cable ON cable_maintenance_logs(cable_route_id);
CREATE INDEX idx_maintenance_logs_segment ON cable_maintenance_logs(network_segment_id);

-- ONU DEVICES TABLE
CREATE TABLE onu_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    serial_number VARCHAR(100),
    mac_address MACADDR,
    ip_address INET,
    status device_status DEFAULT 'online',
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    odp_id UUID REFERENCES odps(id) ON DELETE SET NULL,
    ssid VARCHAR(50),
    password VARCHAR(100),
    model VARCHAR(100),
    firmware_version VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE (mikrotik_id, serial_number)
);

CREATE INDEX idx_onu_mikrotik ON onu_devices(mikrotik_id);
CREATE INDEX idx_onu_customer ON onu_devices(customer_id);
CREATE INDEX idx_onu_odp ON onu_devices(odp_id);
CREATE INDEX idx_onu_status ON onu_devices(status);

CREATE TRIGGER set_updated_at_onu_devices
    BEFORE UPDATE ON onu_devices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

