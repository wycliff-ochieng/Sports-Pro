-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd


-- Table to store metadata about associated media
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL, -- The ID of the workout or exercise it belongs to
    parent_type VARCHAR(50) NOT NULL CHECK (parent_type IN ('WORKOUT', 'EXERCISE')),
    storage_provider VARCHAR(50) NOT NULL DEFAULT 'S3',
    bucket_name VARCHAR(255) NOT NULL,
    object_key VARCHAR(1024) NOT NULL UNIQUE, -- The full path/key of the file in the S3 bucket
    file_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_parent ON media(parent_id, parent_type);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
