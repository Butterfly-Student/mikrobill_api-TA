-- +goose Up
-- +goose StatementBegin

CREATE TYPE odp_status AS ENUM ('active', 'maintenance', 'inactive');
CREATE TYPE connection_type AS ENUM ('fiber', 'copper', 'wireless', 'microwave');
CREATE TYPE cable_capacity AS ENUM ('100M', '1G', '10G', '100G');

-- ODPs TABLE (Optical Distribution Point)
CREATE TABLE odps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    parent_odp_id UUID REFERENCES odps(id) ON DELETE SET NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    address TEXT,
    capacity INTEGER DEFAULT 64,
    used_ports INTEGER DEFAULT 0,
    is_pole BOOLEAN DEFAULT false,
    status odp_status DEFAULT 'active',
    installation_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_odps_location ON odps(latitude, longitude);
CREATE INDEX idx_odps_status ON odps(status);
CREATE INDEX idx_odps_parent ON odps(parent_odp_id);

-- ODP CONNECTIONS TABLE
CREATE TABLE odp_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_odp_id UUID NOT NULL REFERENCES odps(id) ON DELETE CASCADE,
    to_odp_id UUID NOT NULL REFERENCES odps(id) ON DELETE CASCADE,
    connection_type connection_type DEFAULT 'fiber',
    cable_length DECIMAL(8,2),
    cable_capacity cable_capacity DEFAULT '1G',
    status odp_status DEFAULT 'active',
    installation_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (from_odp_id, to_odp_id)
);

CREATE INDEX idx_odp_conn_from ON odp_connections(from_odp_id);
CREATE INDEX idx_odp_conn_to ON odp_connections(to_odp_id);

CREATE TRIGGER set_updated_at_odps
    BEFORE UPDATE ON odps
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_odp_connections
    BEFORE UPDATE ON odp_connections
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

