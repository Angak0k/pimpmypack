CREATE TABLE inventory_images (
    item_id INTEGER PRIMARY KEY REFERENCES inventory(id) ON DELETE CASCADE,
    image_data BYTEA NOT NULL,
    mime_type VARCHAR(50) NOT NULL DEFAULT 'image/jpeg',
    file_size INTEGER NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
