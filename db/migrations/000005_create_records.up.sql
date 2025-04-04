CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    release_year INTEGER,
    catalog_number VARCHAR(100),
    condition VARCHAR(50),
    notes TEXT,
    cover_image_url VARCHAR(255),
    storage_location VARCHAR(255),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label_id INTEGER REFERENCES labels(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_records_user_id ON records(user_id);
CREATE INDEX idx_records_title ON records(title);