-- +goose Up
CREATE TABLE team_members(
    team_id UUID ,
    role VARCHAR(100) NOT NULL,
    joinedat TIMESTAMP, 
    user_id UUID PRIMARY KEY, --conceptually linked to user service.profile.uuid
    CONSTRAINT team_constraint FOREIGN KEY(team_id)
    REFERENCES teams(id)
);

-- +goose Down