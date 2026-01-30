-- migrations/XXXXXXXX_create_roles_table.sql

-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL -- e.g., 'player', 'coach', 'manager'
);

-- Seed the table with your initial roles
INSERT INTO roles (name) VALUES ('player'), ('coach'), ('manager');

-- +goose Down
DROP TABLE roles;