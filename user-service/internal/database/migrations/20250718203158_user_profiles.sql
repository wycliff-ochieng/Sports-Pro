-- +goose Up
CREATE TABLE profiles (
    user_id INT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL DEFAULT '', -- Default is important!
    last_name VARCHAR(100) NOT NULL DEFAULT '',  -- Default is important!
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS profiles;
