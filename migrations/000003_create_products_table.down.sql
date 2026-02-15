-- Drop products table and related objects
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TABLE IF EXISTS products;
