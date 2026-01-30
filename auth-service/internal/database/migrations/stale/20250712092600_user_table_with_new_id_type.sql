-- +goose Up
CREATE TABLE IF NOT EXISTS Users(
    UserID SERIAL PRIMARY KEY,
    FirstName VARCHAR(100) NOT NULL,
    LastName VARCHAR(100) NOT NULL,
    Email VARCHAR(100) UNIQUE NOT NULL,
    Password TEXT NOT NULL,
    PhoneNumber  BIGINT ,
    DateOfBirth DATE,
    Metadata JSONB,
    CreatedAt TIMESTAMP NOT NULL DEFAULT NOW(),
    UpdatedAT TIMESTAMP  NOT NULL  DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS Users;