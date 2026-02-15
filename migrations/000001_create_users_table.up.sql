-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    phone VARCHAR(20),
    
    -- User hierarchy and permissions
    upline_id UUID REFERENCES users(id),
    level INTEGER NOT NULL DEFAULT 1, -- 1=Reseller, 2=Agent, 3=Master, 4=Admin
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    
    -- Financial information
    balance DECIMAL(19, 4) NOT NULL DEFAULT 0.0000,
    credit_limit DECIMAL(19, 4) DEFAULT 0.0000,
    markup_percentage DECIMAL(5, 2) DEFAULT 0.00, -- Markup percentage for pricing
    
    -- Business settings
    allow_debt BOOLEAN DEFAULT false,
    max_daily_transaction DECIMAL(19, 4) DEFAULT 999999999.0000,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_upline_id ON users(upline_id);
CREATE INDEX idx_users_level ON users(level);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
