-- +goose Up
-- +goose StatementBegin

CREATE TYPE collector_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE payment_method AS ENUM ('cash', 'transfer', 'other');
CREATE TYPE collector_payment_status AS ENUM ('completed', 'pending', 'cancelled');

-- COLLECTORS TABLE
CREATE TABLE collectors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    address TEXT,
    status collector_status DEFAULT 'active',
    commission_rate DECIMAL(5,2) DEFAULT 5.00,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    
    UNIQUE (phone)
);

CREATE INDEX idx_collectors_phone ON collectors(phone);
CREATE INDEX idx_collectors_status ON collectors(status);

-- COLLECTOR PAYMENTS TABLE
CREATE TABLE collector_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    payment_amount DECIMAL(15,2) NOT NULL,
    commission_amount DECIMAL(15,2) NOT NULL,
    payment_method payment_method DEFAULT 'cash',
    payment_date TIMESTAMPTZ DEFAULT now(),
    notes TEXT,
    status collector_payment_status DEFAULT 'completed',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_cp_collector ON collector_payments(collector_id);
CREATE INDEX idx_cp_customer ON collector_payments(customer_id);
CREATE INDEX idx_cp_invoice ON collector_payments(invoice_id);
CREATE INDEX idx_cp_date ON collector_payments(payment_date);

-- EXPENSES TABLE
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description TEXT NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    category VARCHAR(100) NOT NULL,
    expense_date DATE NOT NULL,
    payment_method VARCHAR(50),
    notes TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_expenses_date ON expenses(expense_date);
CREATE INDEX idx_expenses_category ON expenses(category);
CREATE INDEX idx_expenses_created_by ON expenses(created_by);

CREATE TRIGGER set_updated_at_collectors
    BEFORE UPDATE ON collectors
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_collector_payments
    BEFORE UPDATE ON collector_payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at_expenses
    BEFORE UPDATE ON expenses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose StatementEnd

