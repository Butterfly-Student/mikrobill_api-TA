-- -- +goose Up
-- -- +goose StatementBegin

-- -- INVOICES TABLE
-- CREATE TABLE invoices (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
--     customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
--     package_id UUID NOT NULL REFERENCES packages(id) ON DELETE RESTRICT,
--     invoice_number VARCHAR(50) NOT NULL,
--     amount DECIMAL(10,2) NOT NULL,
--     due_date DATE NOT NULL,
--     status invoice_status DEFAULT 'unpaid',
--     invoice_type invoice_type DEFAULT 'monthly',
--     package_name VARCHAR(100),
--     description TEXT,
--     payment_date TIMESTAMPTZ,
--     payment_method VARCHAR(50),
--     payment_gateway VARCHAR(50),
--     payment_token VARCHAR(255),
--     payment_url TEXT,
--     payment_status payment_status DEFAULT 'pending',
--     notes TEXT,
--     created_at TIMESTAMPTZ DEFAULT now(),
--     updated_at TIMESTAMPTZ DEFAULT now(),
    
--     UNIQUE (mikrotik_id, invoice_number)
-- );

-- CREATE INDEX idx_invoices_mikrotik ON invoices(mikrotik_id);
-- CREATE INDEX idx_invoices_customer ON invoices(customer_id);
-- CREATE INDEX idx_invoices_status ON invoices(status);
-- CREATE INDEX idx_invoices_due_date ON invoices(due_date);
-- CREATE INDEX idx_invoices_package ON invoices(package_id);

-- -- PAYMENTS TABLE
-- CREATE TABLE payments (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
--     amount DECIMAL(10,2) NOT NULL,
--     payment_date TIMESTAMPTZ DEFAULT now(),
--     payment_method VARCHAR(50) NOT NULL,
--     reference_number VARCHAR(100),
--     notes TEXT,
--     created_at TIMESTAMPTZ DEFAULT now()
-- );

-- CREATE INDEX idx_payments_invoice ON payments(invoice_id);
-- CREATE INDEX idx_payments_date ON payments(payment_date);

-- -- PAYMENT GATEWAY TRANSACTIONS
-- CREATE TABLE payment_gateway_transactions (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
--     gateway VARCHAR(50) NOT NULL,
--     order_id VARCHAR(100) NOT NULL,
--     payment_url TEXT,
--     token VARCHAR(255),
--     amount DECIMAL(10,2) NOT NULL,
--     status payment_status DEFAULT 'pending',
--     payment_type VARCHAR(50),
--     fraud_status VARCHAR(50),
--     created_at TIMESTAMPTZ DEFAULT now(),
--     updated_at TIMESTAMPTZ DEFAULT now()
-- );

-- CREATE INDEX idx_pgt_invoice ON payment_gateway_transactions(invoice_id);
-- CREATE INDEX idx_pgt_order ON payment_gateway_transactions(order_id);
-- CREATE INDEX idx_pgt_status ON payment_gateway_transactions(status);

-- CREATE TRIGGER set_updated_at_invoices
--     BEFORE UPDATE ON invoices
--     FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- CREATE TRIGGER set_updated_at_pgt
--     BEFORE UPDATE ON payment_gateway_transactions
--     FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -- +goose StatementEnd