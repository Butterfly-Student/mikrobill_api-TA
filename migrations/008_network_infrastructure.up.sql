-- +goose Up
-- +goose StatementBegin

CREATE TYPE segment_type AS ENUM ('Backbone', 'Distribution', 'Access');
CREATE TYPE segment_status AS ENUM ('active', 'maintenance', 'damaged', 'inactive');

-- CABLE ROUTES TABLE
CREATE TABLE cable_routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    odp_id UUID NOT NULL REFERENCES odps(id) ON DELETE CASCADE,
    cable_length DECIMAL(8,2),
    cable_type VARCHAR(50) DEFAULT 'Fiber Optic',
    installation_date DATE,
    status cable_status DEFAULT 'connected',
    port_number INTEGER,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_cable_routes_customer ON cable_routes(customer_id);
CREATE INDEX idx_cable_routes_odp ON cable_routes(odp_id);

-- NETWORK SEGMENTS TABLE
CREATE TABLE network_segments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    start_odp_id UUID NOT NULL REFERENCES odps(id) ON DELETE CASCADE,
    end_odp_id UUID REFERENCES odps(id) ON DELETE CASCADE,
    segment_type segment_type DEFAULT 'Backbone',
    cable_length DECIMAL(10,2),
    status segment_status DEFAULT 'active',
    installation_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_network_segments_start ON network_segments(start_odp_id);
CREATE INDEX idx_network_segments_end ON network_segments(end_odp_id);

CREATE TRIGGER set_updated_at_cable_routes
    BEFORE UPDATE ON cable_routes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_network_segments
    BEFORE UPDATE ON network_segments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

