-- Миграция 002: Add support for arbitrary metadata
CREATE TABLE IF NOT EXISTS data_metadata (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    data_id UUID REFERENCES user_data(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value_type SMALLINT NOT NULL, -- 1: text, 2: number, 3: boolean, 4: binary
    text_value TEXT,
    number_value NUMERIC,
    boolean_value BOOLEAN,
    binary_value BYTEA,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_data_metadata_data_id ON data_metadata(data_id);
CREATE INDEX IF NOT EXISTS idx_data_metadata_key ON data_metadata(key);

-- Добавляем поддержку категорий для организации метаданных
CREATE TABLE IF NOT EXISTS metadata_categories (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS data_metadata_categories (
    metadata_id INTEGER REFERENCES data_metadata(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES metadata_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (metadata_id, category_id)
);