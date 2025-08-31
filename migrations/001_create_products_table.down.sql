DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX CONCURRENTLY IF EXISTS idx_products_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_price;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_category;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_brand;
DROP TABLE IF EXISTS products;