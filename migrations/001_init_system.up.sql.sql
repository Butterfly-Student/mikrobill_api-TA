-- +goose Up
-- +goose StatementBegin

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ENUM Types
CREATE TYPE user_status AS ENUM ('active','inactive','suspended','locked');
CREATE TYPE user_role AS ENUM ('superadmin','admin','technician','sales','cs','finance','viewer');
CREATE TYPE mikrotik_status AS ENUM ('online', 'offline', 'error', 'maintenance');
CREATE TYPE profile_type AS ENUM ('pppoe', 'hotspot');
CREATE TYPE invoice_status AS ENUM ('unpaid', 'paid', 'overdue', 'cancelled');
CREATE TYPE invoice_type AS ENUM ('monthly', 'voucher', 'installation', 'other');
CREATE TYPE payment_status AS ENUM ('pending', 'success', 'failed', 'expired');
CREATE TYPE voucher_status AS ENUM ('active', 'used', 'expired', 'cancelled');
CREATE TYPE agent_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE agent_transaction_type AS ENUM ('deposit', 'withdrawal', 'voucher_sale', 'monthly_payment', 'commission', 'balance_request');
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'cancelled');
CREATE TYPE collector_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE payment_method AS ENUM ('cash', 'transfer', 'other');
CREATE TYPE collector_payment_status AS ENUM ('completed', 'pending', 'cancelled');

-- +goose StatementEnd