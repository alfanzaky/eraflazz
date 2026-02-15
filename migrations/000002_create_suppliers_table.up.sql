-- Create suppliers table
CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL, -- Supplier code: DIGIFLAZZ, VIP, etc.
    
    -- API Configuration
    api_url VARCHAR(255) NOT NULL,
    api_key VARCHAR(255),
    api_secret VARCHAR(255),
    api_username VARCHAR(100),
    api_password VARCHAR(255),
    
    -- Supplier status and settings
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 1, -- Lower number = higher priority
    timeout_seconds INTEGER DEFAULT 30,
    retry_attempts INTEGER DEFAULT 3,
    
    -- Financial information
    balance DECIMAL(19, 4) DEFAULT 0.0000, -- Balance at supplier
    min_balance_threshold DECIMAL(19, 4) DEFAULT 0.0000,
    
    -- Performance metrics
    success_rate DECIMAL(5, 2) DEFAULT 100.00, -- Percentage
    avg_response_time_ms INTEGER DEFAULT 1000,
    total_transactions INTEGER DEFAULT 0,
    failed_transactions INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_checked_at TIMESTAMP WITH TIME ZONE,
    last_success_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_suppliers_code ON suppliers(code);
CREATE INDEX idx_suppliers_is_active ON suppliers(is_active);
CREATE INDEX idx_suppliers_priority ON suppliers(priority);
CREATE INDEX idx_suppliers_success_rate ON suppliers(success_rate);

-- Trigger for updated_at
CREATE TRIGGER update_suppliers_updated_at 
    BEFORE UPDATE ON suppliers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
