-- -- +goose Up
-- -- +goose StatementBegin

-- -- AGENTS TABLE
-- CREATE TABLE agents (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     username VARCHAR(100) NOT NULL,
--     name VARCHAR(255) NOT NULL,
--     phone VARCHAR(20) NOT NULL,
--     email VARCHAR(255),
--     password VARCHAR(255) NOT NULL,
--     address TEXT,
--     status agent_status DEFAULT 'active',
--     commission_rate DECIMAL(5,2) DEFAULT 5.00,
--     created_at TIMESTAMPTZ DEFAULT now(),
--     updated_at TIMESTAMPTZ DEFAULT now(),
    
--     UNIQUE (username),
--     UNIQUE (phone)
-- );

-- CREATE INDEX idx_agents_username ON agents(username);
-- CREATE INDEX idx_agents_phone ON agents(phone);
-- CREATE INDEX idx_agents_status ON agents(status);

-- -- AGENT BALANCES TABLE
-- CREATE TABLE agent_balances (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
--     balance DECIMAL(15,2) DEFAULT 0.00,
--     last_updated TIMESTAMPTZ DEFAULT now(),
    
--     UNIQUE (agent_id)
-- );

-- CREATE INDEX idx_agent_balances_agent ON agent_balances(agent_id);

-- -- AGENT TRANSACTIONS TABLE
-- CREATE TABLE agent_transactions (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
--     transaction_type agent_transaction_type NOT NULL,
--     amount DECIMAL(15,2) NOT NULL,
--     description TEXT,
--     reference_id VARCHAR(100),
--     status transaction_status DEFAULT 'completed',
--     created_at TIMESTAMPTZ DEFAULT now()
-- );

-- CREATE INDEX idx_agent_trans_agent ON agent_transactions(agent_id);
-- CREATE INDEX idx_agent_trans_type ON agent_transactions(transaction_type);
-- CREATE INDEX idx_agent_trans_date ON agent_transactions(created_at);

-- CREATE TRIGGER set_updated_at_agents
--     BEFORE UPDATE ON agents
--     FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -- +goose StatementEnd