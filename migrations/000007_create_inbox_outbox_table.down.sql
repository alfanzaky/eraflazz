-- Drop inbox and outbox tables and related objects
DROP TRIGGER IF EXISTS update_inbox_updated_at ON inbox;
DROP TRIGGER IF EXISTS update_outbox_updated_at ON outbox;
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS inbox;
