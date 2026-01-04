-- +goose Up
-- +goose StatementBegin

-- APPLICATION SETTINGS TABLE
CREATE TABLE app_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category VARCHAR(50) NOT NULL, -- 'general', 'billing', 'network', 'notification', 'integration'
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    setting_type VARCHAR(20) DEFAULT 'string', -- 'string', 'number', 'boolean', 'json'
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE (category, setting_key)
);

CREATE INDEX idx_app_settings_category ON app_settings(category);
CREATE INDEX idx_app_settings_key ON app_settings(setting_key);

-- COMPANY PROFILE TABLE
CREATE TABLE company_profile (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_name VARCHAR(255) NOT NULL,
    company_address TEXT,
    company_phone VARCHAR(20),
    company_email VARCHAR(255),
    company_website VARCHAR(255),
    
    -- Logo & branding
    logo_url VARCHAR(255),
    favicon_url VARCHAR(255),
    primary_color VARCHAR(7) DEFAULT '#3B82F6',
    secondary_color VARCHAR(7) DEFAULT '#1E40AF',
    
    -- Invoice settings
    invoice_prefix VARCHAR(10) DEFAULT 'INV',
    invoice_start_number INTEGER DEFAULT 1000,
    invoice_terms TEXT,
    invoice_footer TEXT,
    
    -- Tax settings
    default_tax_rate DECIMAL(5,2) DEFAULT 11.00,
    tax_id VARCHAR(50),
    
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- NOTIFICATION TEMPLATES
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_name VARCHAR(100) NOT NULL,
    template_type VARCHAR(50) NOT NULL, -- 'email', 'sms', 'whatsapp'
    subject VARCHAR(255),
    content TEXT NOT NULL,
    variables TEXT[], -- Array of variables used in template
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE (template_name, template_type)
);

CREATE INDEX idx_notification_templates_type ON notification_templates(template_type);
CREATE INDEX idx_notification_templates_active ON notification_templates(is_active);

-- Insert default settings
INSERT INTO app_settings (category, setting_key, setting_value, description) VALUES
('general', 'app_name', 'ISP Management System', 'Application name'),
('general', 'timezone', 'Asia/Jakarta', 'Default timezone'),
('general', 'currency', 'IDR', 'Default currency'),
('general', 'date_format', 'DD/MM/YYYY', 'Date format'),
('billing', 'auto_generate_invoice', 'true', 'Auto generate invoice on billing day'),
('billing', 'invoice_due_days', '7', 'Number of days until invoice due'),
('billing', 'late_fee_percentage', '2', 'Late fee percentage'),
('billing', 'grace_period_days', '3', 'Grace period before suspension'),
('notification', 'invoice_reminder_days', '3,1,0', 'Days before due date to send reminder'),
('notification', 'enable_email_notifications', 'true', 'Enable email notifications'),
('notification', 'enable_sms_notifications', 'false', 'Enable SMS notifications');

INSERT INTO company_profile (company_name, company_address, company_phone, company_email) VALUES
('Your ISP Company', 'Jl. Contoh No. 123', '021-1234567', 'admin@yourisp.com');

-- Default notification templates
INSERT INTO notification_templates (template_name, template_type, subject, content, variables) VALUES
('invoice_created', 'email', 'Invoice {{invoice_number}} telah dibuat', 
'Yth. {{customer_name}},\n\nInvoice {{invoice_number}} dengan jumlah {{invoice_amount}} telah dibuat. Jatuh tempo: {{due_date}}.\n\nTerima kasih.', 
ARRAY['customer_name', 'invoice_number', 'invoice_amount', 'due_date']),
('payment_received', 'email', 'Pembayaran diterima', 
'Yth. {{customer_name}},\n\nPembayaran untuk invoice {{invoice_number}} sebesar {{payment_amount}} telah diterima.\n\nTerima kasih.', 
ARRAY['customer_name', 'invoice_number', 'payment_amount']),
('account_suspended', 'email', 'Akun Anda telah ditangguhkan', 
'Yth. {{customer_name}},\n\nAkun Anda telah ditangguhkan karena pembayaran terlambat. Silakan lakukan pembayaran untuk mengaktifkan kembali layanan.\n\nTerima kasih.', 
ARRAY['customer_name']);

CREATE TRIGGER set_updated_at_app_settings
    BEFORE UPDATE ON app_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_company_profile
    BEFORE UPDATE ON company_profile
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_notification_templates
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd