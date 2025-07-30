-- +goose Up
CREATE TABLE user_profiles (
    userid INT PRIMARY KEY,
    firstname VARCHAR(100) NOT NULL DEFAULT '', -- Default is important!
    lastname VARCHAR(100) NOT NULL DEFAULT '',  -- Default is important!
    email VARCHAR(255) NOT NULL UNIQUE,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS user_profiles;
