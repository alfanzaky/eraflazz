-- Create products table
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL, -- Product code: T10, S5, PLN20, etc.
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Product categorization
    category VARCHAR(50) NOT NULL, -- PULSA, DATA, PLN, PDAM, BPJS, GAME, etc.
    provider VARCHAR(50) NOT NULL, -- TELKOMSEL, INDOSAT, XL, etc.
    type VARCHAR(20) NOT NULL, -- PREPAID, POSTPAID, VOUCHER
    
    -- Pricing information
    base_price DECIMAL(19, 4) NOT NULL DEFAULT 0.0000, -- Base price from supplier
    selling_price DECIMAL(19, 4) NOT NULL DEFAULT 0.0000, -- Default selling price
    min_price DECIMAL(19, 4) DEFAULT 0.0000, -- Minimum allowed price
    
    -- Product specifications
    nominal DECIMAL(10, 0), -- For pulsa/data: 10000, 5000, etc.
    validity_period VARCHAR(50), -- e.g., "30 Hari", "7 Hari"
    
    -- Status and availability
    is_active BOOLEAN DEFAULT true,
    is_unlimited_stock BOOLEAN DEFAULT false,
    stock_quantity INTEGER DEFAULT 0,
    
    -- Business rules
    allow_markup BOOLEAN DEFAULT true,
    max_markup_percentage DECIMAL(5, 2) DEFAULT 100.00,
    min_transaction_amount DECIMAL(19, 4) DEFAULT 1.0000,
    max_transaction_amount DECIMAL(19, 4) DEFAULT 999999999.0000,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_products_code ON products(code);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_provider ON products(provider);
CREATE INDEX idx_products_type ON products(type);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_base_price ON products(base_price);

-- Trigger for updated_at
CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
