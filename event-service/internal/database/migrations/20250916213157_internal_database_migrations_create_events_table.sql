-- +goose Up
-- +goose StatementBegin
CREATE TABLE events(
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_title VARCHAR(50),
    event_type VARCHAR(20),
    location VARCHAR(20),
    start_time TIMESTAMP DEFAULT NOW(),
    end_time TIMESTAMP DEFAULT NOW()
) ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd