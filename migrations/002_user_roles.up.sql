-- +goose Up
-- +goose StatementBegin

-- ENUM types untuk users dan roles
CREATE TYPE user_status AS ENUM ('active','inactive','suspended','locked');
CREATE TYPE user_role AS ENUM ('superadmin','admin','technician','sales','cs','finance','viewer');

-- Tabel roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tabel users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    encrypted_password TEXT NOT NULL,
    
    -- Profile
    fullname VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar TEXT,
    
    -- Role & Status
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    user_role user_role NOT NULL DEFAULT 'viewer',
    status user_status NOT NULL DEFAULT 'active',
    
    -- Security
    last_login TIMESTAMPTZ,
    last_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ,
    force_password_change BOOLEAN DEFAULT false,
    
    -- Two Factor
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret TEXT,
    
    -- API Access
    api_token TEXT UNIQUE,
    api_token_expires_at TIMESTAMPTZ,
    
    -- Metadata
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role_id);
CREATE INDEX idx_users_status ON users (status);
CREATE INDEX idx_users_api_token ON users (api_token);

CREATE TRIGGER set_updated_at_users
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_roles
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Insert default roles
INSERT INTO roles (name, display_name, description, is_system) VALUES
('superadmin', 'Super Administrator', 'Full system access', true),
('admin', 'Administrator', 'Administrative access', true),
('technician', 'Technician', 'Technical operations', true),
('sales', 'Sales', 'Sales operations', true),
('cs', 'Customer Service', 'Customer support', true),
('finance', 'Finance', 'Financial operations', true),
('viewer', 'Viewer', 'Read-only access', true);

-- +goose StatementEnd

