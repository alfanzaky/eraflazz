-- Create api_clients table for H2H authentication
CREATE TABLE api_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(100) UNIQUE NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    secret VARCHAR(255) NOT NULL,
    
    -- Security settings
    ip_whitelist TEXT[], -- Array of allowed IP addresses
    is_active BOOLEAN DEFAULT true,
    
    -- Rate limiting
    max_requests_per_minute INTEGER DEFAULT 60,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX idx_api_clients_client_id ON api_clients(client_id);
CREATE INDEX idx_api_clients_api_key ON api_clients(api_key);
CREATE INDEX idx_api_clients_is_active ON api_clients(is_active);
CREATE INDEX idx_api_clients_created_at ON api_clients(created_at);

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_api_clients_updated_at 
    BEFORE UPDATE ON api_clients 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
