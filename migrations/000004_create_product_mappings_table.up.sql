-- Create product_mappings table (many-to-many relationship between products and suppliers)
CREATE TABLE product_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    
    -- Supplier-specific product code
    supplier_product_code VARCHAR(50) NOT NULL, -- Code used by supplier
    
    -- Supplier-specific pricing
    supplier_price DECIMAL(19, 4) NOT NULL DEFAULT 0.0000, -- Price from this supplier
    additional_fee DECIMAL(19, 4) DEFAULT 0.0000, -- Additional fees (admin, etc.)
    
    -- Priority and availability
    priority INTEGER DEFAULT 1, -- Priority for this supplier (1=highest)
    is_active BOOLEAN DEFAULT true,
    
    -- Stock information from supplier
    stock_status VARCHAR(20) DEFAULT 'UNKNOWN', -- AVAILABLE, OUT_OF_STOCK, UNKNOWN
    last_stock_check TIMESTAMP WITH TIME ZONE,
    
    -- Performance metrics for this mapping
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    last_success_at TIMESTAMP WITH TIME ZONE,
    last_failure_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure unique combination of product and supplier
    UNIQUE(product_id, supplier_id)
);

-- Indexes
CREATE INDEX idx_product_mappings_product_id ON product_mappings(product_id);
CREATE INDEX idx_product_mappings_supplier_id ON product_mappings(supplier_id);
CREATE INDEX idx_product_mappings_priority ON product_mappings(priority);
CREATE INDEX idx_product_mappings_is_active ON product_mappings(is_active);
CREATE INDEX idx_product_mappings_supplier_price ON product_mappings(supplier_price);

-- Trigger for updated_at
CREATE TRIGGER update_product_mappings_updated_at 
    BEFORE UPDATE ON product_mappings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
