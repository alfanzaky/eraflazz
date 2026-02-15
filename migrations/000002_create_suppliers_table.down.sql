-- Drop suppliers table and related objects
DROP TRIGGER IF EXISTS update_suppliers_updated_at ON suppliers;
DROP TABLE IF EXISTS suppliers;
