--- Миграция 001: Initial schema for Password Manager
--- Создание базовых таблиц пользователей и данных
DROP INDEX IF EXISTS idx_user_data_updated_at;
DROP INDEX IF EXISTS idx_user_data_user_id;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS user_data;
DROP TABLE IF EXISTS users;