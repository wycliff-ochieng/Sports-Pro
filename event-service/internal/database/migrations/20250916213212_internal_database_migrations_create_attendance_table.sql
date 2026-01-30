-- +goose Up
-- +goose StatementBegin
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp"

CREATE TABLE attendance (
    team_id uuid NOT NULL,
    event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    event_status VARCHAR(20),
    updated_at TIMESTAMP DEFAULT NOW()
) ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd




