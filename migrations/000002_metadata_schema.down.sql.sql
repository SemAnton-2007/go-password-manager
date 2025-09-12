--- Миграция 002: Add support for arbitrary metadata
DROP TABLE IF EXISTS data_metadata_categories;
DROP TABLE IF EXISTS metadata_categories;
DROP INDEX IF EXISTS idx_data_metadata_key;
DROP INDEX IF EXISTS idx_data_metadata_data_id;
DROP TABLE IF EXISTS data_metadata;