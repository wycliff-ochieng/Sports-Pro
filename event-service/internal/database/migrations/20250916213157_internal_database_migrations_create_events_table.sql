-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"


CREATE TEBALE IF NOT EXISTS attendance (
    team_id uuid
    event_id uuid PRIMARY KEY DEFAULT uuid_generate_v4
    user_id uuid
    event_status VARCHAR(20)
) ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
