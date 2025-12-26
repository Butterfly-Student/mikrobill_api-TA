-- +goose Up
-- +goose StatementBegin

CREATE TYPE request_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE notification_type AS ENUM ('voucher_sold', 'payment_received', 'balance_updated', 'request_approved', 'request_rejected');

-- AGENT VOUCHER SALES TABLE
CREATE TABLE agent_voucher_sales (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
    voucher_code VARCHAR(100) NOT NULL,
    package_id VARCHAR(50) NOT NULL,
    package_name VARCHAR(100) NOT NULL,
    customer_phone VARCHAR(20),
    customer_name VARCHAR(255),
    price DECIMAL(10,2) NOT NULL,
    agent_price DECIMAL(10,2) DEFAULT 0.00,
    commission DECIMAL(10,2) DEFAULT 0.00,
    commission_amount DECIMAL(10,2) DEFAULT 0.00,
    status voucher_status DEFAULT 'active',
    sold_at TIMESTAMPTZ DEFAULT now(),
    used_at TIMESTAMPTZ,
    notes TEXT,
    
    UNIQUE (mikrotik_id, voucher_code)
);

CREATE INDEX idx_avs_agent ON agent_voucher_sales(agent_id);
CREATE INDEX idx_avs_mikrotik ON agent_voucher_sales(mikrotik_id);
CREATE INDEX idx_avs_code ON agent_voucher_sales(voucher_code);
CREATE INDEX idx_avs_status ON agent_voucher_sales(status);

-- AGENT BALANCE REQUESTS TABLE
CREATE TABLE agent_balance_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL,
    status request_status DEFAULT 'pending',
    admin_notes TEXT,
    requested_at TIMESTAMPTZ DEFAULT now(),
    processed_at TIMESTAMPTZ,
    processed_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_abr_agent ON agent_balance_requests(agent_id);
CREATE INDEX idx_abr_status ON agent_balance_requests(status);

-- AGENT MONTHLY PAYMENTS TABLE
CREATE TABLE agent_monthly_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    payment_amount DECIMAL(15,2) NOT NULL,
    commission_amount DECIMAL(15,2) DEFAULT 0.00,
    payment_method VARCHAR(50) DEFAULT 'cash',
    notes TEXT,
    status transaction_status DEFAULT 'completed',
    paid_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_amp_agent ON agent_monthly_payments(agent_id);
CREATE INDEX idx_amp_customer ON agent_monthly_payments(customer_id);
CREATE INDEX idx_amp_invoice ON agent_monthly_payments(invoice_id);

-- AGENT PAYMENTS TABLE
CREATE TABLE agent_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(50) DEFAULT 'cash',
    notes TEXT,
    status transaction_status DEFAULT 'completed',
    paid_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_ap_agent ON agent_payments(agent_id);
CREATE INDEX idx_ap_customer ON agent_payments(customer_id);
CREATE INDEX idx_ap_invoice ON agent_payments(invoice_id);

-- AGENT NOTIFICATIONS TABLE
CREATE TABLE agent_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    notification_type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_an_agent ON agent_notifications(agent_id);
CREATE INDEX idx_an_read ON agent_notifications(is_read);

-- +goose StatementEnd


-- +goose StatementEnd