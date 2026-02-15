-- Create transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trx_code VARCHAR(50) UNIQUE NOT NULL, -- Format: TRX-YYYYMMDD-XXXX
    user_id UUID NOT NULL REFERENCES users(id),
    product_id UUID NOT NULL REFERENCES products(id),
    supplier_id UUID REFERENCES suppliers(id), -- Will be filled after routing
    
    -- Transaction details
    destination_number VARCHAR(50) NOT NULL, -- Phone number, customer ID, etc.
    product_code VARCHAR(20) NOT NULL, -- Product code at time of transaction
    
    -- Pricing information (snapshot to prevent changes)
    hpp DECIMAL(19, 4) NOT NULL DEFAULT 0.0000, -- Cost price from supplier
    selling_price DECIMAL(19, 4) NOT NULL DEFAULT 0.0000, -- Price charged to user
    admin_fee DECIMAL(19, 4) DEFAULT 0.0000, -- Additional fees
    profit DECIMAL(19, 4) GENERATED ALWAYS AS (selling_price - hpp - admin_fee) STORED,
    
    -- Transaction status
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (
        status IN ('PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'REFUND', 'TIMEOUT')
    ),
    
    -- Supplier response
    serial_number VARCHAR(100), -- SN/token from supplier
    supplier_message TEXT, -- Response message from supplier
    supplier_trx_id VARCHAR(100), -- Transaction ID at supplier
    
    -- Routing information
    routing_attempts INTEGER DEFAULT 0, -- Number of supplier attempts
    final_supplier_id UUID REFERENCES suppliers(id), -- Final supplier that succeeded
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE, -- When processing started
    completed_at TIMESTAMP WITH TIME ZONE, -- When transaction completed
    
    -- Additional metadata
    user_ip INET,
    user_agent TEXT,
    api_endpoint VARCHAR(100), -- Which API endpoint was used
    notes TEXT -- Internal notes
);

-- Indexes for performance
CREATE INDEX idx_transactions_trx_code ON transactions(trx_code);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_product_id ON transactions(product_id);
CREATE INDEX idx_transactions_supplier_id ON transactions(supplier_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_destination_number ON transactions(destination_number);
CREATE INDEX idx_transactions_completed_at ON transactions(completed_at);

-- Partial indexes for common queries
CREATE INDEX idx_transactions_pending ON transactions(created_at) WHERE status = 'PENDING';
CREATE INDEX idx_transactions_processing ON transactions(created_at) WHERE status = 'PROCESSING';
CREATE INDEX idx_transactions_success ON transactions(created_at) WHERE status = 'SUCCESS';
CREATE INDEX idx_transactions_failed ON transactions(created_at) WHERE status = 'FAILED';

-- Trigger for updated_at
CREATE TRIGGER update_transactions_updated_at 
    BEFORE UPDATE ON transactions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
