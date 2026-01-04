-- +goose Up
-- +goose StatementBegin

CREATE TABLE mikrotik (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    host INET NOT NULL,
    port INTEGER NOT NULL DEFAULT 8728,
    api_username TEXT NOT NULL,
    api_encrypted_password TEXT,
    keepalive BOOLEAN DEFAULT true,
    timeout INTEGER DEFAULT 300000,
    location VARCHAR(100),
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    status mikrotik_status NOT NULL DEFAULT 'offline',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (host, port)
);

CREATE INDEX idx_mikrotik_status ON mikrotik(status);
CREATE INDEX idx_mikrotik_is_active ON mikrotik(is_active);
CREATE INDEX idx_mikrotik_host ON mikrotik(host);

CREATE TRIGGER set_updated_at_mikrotik
    BEFORE UPDATE ON mikrotik
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- Insert sample MikroTik device for testing
INSERT INTO mikrotik (name, host, port, api_username, api_encrypted_password, location, status)
VALUES 
    ('MikroTik Main', '192.168.100.1', 8728, 'admin', 'r00t', 'Data Center', 'online')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd