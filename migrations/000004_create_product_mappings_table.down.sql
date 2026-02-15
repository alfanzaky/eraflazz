-- Drop product_mappings table and related objects
DROP TRIGGER IF EXISTS update_product_mappings_updated_at ON product_mappings;
DROP TABLE IF EXISTS product_mappings;
