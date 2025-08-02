-- migrations/XXXXXXXX_create_user_roles_table.sql

-- +goose Up
CREATE TABLE user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (user_id, role_id), -- Ensures a user can't have the same role twice
    FOREIGN KEY (user_id) REFERENCES users(userid) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE user_roles;