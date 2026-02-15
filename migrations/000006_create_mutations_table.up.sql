-- Create mutations table for double-entry accounting
CREATE TABLE mutations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    -- Transaction type and amount
    type VARCHAR(10) NOT NULL CHECK (type IN ('DEBIT', 'CREDIT')), -- DEBIT=money in, CREDIT=money out
    amount DECIMAL(19, 4) NOT NULL,
    
    -- Balance snapshot (crucial for audit trail)
    balance_before DECIMAL(19, 4) NOT NULL,
    balance_after DECIMAL(19, 4) NOT NULL,
    
    -- Transaction reference
    reference_type VARCHAR(20), -- TRANSACTION, DEPOSIT, WITHDRAWAL, COMMISSION, PENALTY
    reference_id UUID, -- ID of related transaction/deposit/etc.
    
    -- Description and metadata
    description TEXT NOT NULL, -- Human-readable description
    notes TEXT, -- Additional internal notes
    
    -- System information
    created_by UUID REFERENCES users(id), -- Who created this mutation
    ip_address INET,
    user_agent TEXT,
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraint: balance_after must equal balance_before +/- amount
    CONSTRAINT check_balance_after CHECK (
        (type = 'DEBIT' AND balance_after = balance_before + amount) OR
        (type = 'CREDIT' AND balance_after = balance_before - amount)
    )
);

-- Indexes
CREATE INDEX idx_mutations_user_id ON mutations(user_id);
CREATE INDEX idx_mutations_type ON mutations(type);
CREATE INDEX idx_mutations_reference_id ON mutations(reference_id);
CREATE INDEX idx_mutations_created_at ON mutations(created_at);
CREATE INDEX idx_mutations_reference_type ON mutations(reference_type);

-- Composite index for balance history queries
CREATE INDEX idx_mutations_user_created ON mutations(user_id, created_at DESC);

-- Index for finding mutations by reference
CREATE INDEX idx_mutations_reference ON mutations(reference_type, reference_id);
