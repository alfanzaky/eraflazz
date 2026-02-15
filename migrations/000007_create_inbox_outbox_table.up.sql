-- Create inbox table for incoming messages
CREATE TABLE inbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Message source
    source VARCHAR(20) NOT NULL, -- WHATSAPP, TELEGRAM, SMS, API
    sender_number VARCHAR(50) NOT NULL, -- Phone number or user ID
    sender_name VARCHAR(100), -- Contact name if available
    
    -- Message content
    message TEXT NOT NULL,
    original_message TEXT, -- Original message before processing
    
    -- Processing information
    user_id UUID REFERENCES users(id), -- User identified from sender
    transaction_id UUID REFERENCES transactions(id), -- Related transaction if any
    
    -- Processing status
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (
        status IN ('PENDING', 'PROCESSING', 'PROCESSED', 'FAILED', 'IGNORED')
    ),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    -- Response
    response_message TEXT,
    response_sent_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    ip_address INET,
    device_info TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create outbox table for outgoing messages
CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Message destination
    destination VARCHAR(20) NOT NULL, -- WHATSAPP, TELEGRAM, SMS, API
    recipient_number VARCHAR(50) NOT NULL,
    recipient_name VARCHAR(100),
    
    -- Message content
    message TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'NOTIFICATION', -- NOTIFICATION, TRANSACTION, ALERT, MARKETING
    
    -- Related entities
    user_id UUID REFERENCES users(id),
    transaction_id UUID REFERENCES transactions(id),
    
    -- Sending status
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (
        status IN ('PENDING', 'SENDING', 'SENT', 'FAILED', 'CANCELLED')
    ),
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    -- Sending results
    sent_at TIMESTAMP WITH TIME ZONE,
    delivery_report TEXT, -- Response from gateway
    external_id VARCHAR(100), -- Message ID from external service
    
    -- Scheduling
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    priority INTEGER DEFAULT 1, -- 1=high, 2=normal, 3=low
    created_by UUID REFERENCES users(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for inbox
CREATE INDEX idx_inbox_source ON inbox(source);
CREATE INDEX idx_inbox_sender_number ON inbox(sender_number);
CREATE INDEX idx_inbox_user_id ON inbox(user_id);
CREATE INDEX idx_inbox_status ON inbox(status);
CREATE INDEX idx_inbox_created_at ON inbox(created_at);

-- Indexes for outbox
CREATE INDEX idx_outbox_destination ON outbox(destination);
CREATE INDEX idx_outbox_recipient_number ON outbox(recipient_number);
CREATE INDEX idx_outbox_user_id ON outbox(user_id);
CREATE INDEX idx_outbox_status ON outbox(status);
CREATE INDEX idx_outbox_scheduled_at ON outbox(scheduled_at);
CREATE INDEX idx_outbox_priority ON outbox(priority);

-- Triggers for updated_at
CREATE TRIGGER update_inbox_updated_at 
    BEFORE UPDATE ON inbox 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_outbox_updated_at 
    BEFORE UPDATE ON outbox 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
