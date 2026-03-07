CREATE TABLE account_banner_images (
    account_id INTEGER PRIMARY KEY REFERENCES account(id) ON DELETE CASCADE,
    image_data BYTEA NOT NULL,
    mime_type VARCHAR(50) NOT NULL DEFAULT 'image/jpeg',
    file_size INTEGER NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE account ADD COLUMN banner_position_y INTEGER NOT NULL DEFAULT 50 CHECK (banner_position_y BETWEEN 0 AND 100);
