-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- Junction table to define the contents of a workout
CREATE TABLE workout_exercises (
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    sequence INT NOT NULL, -- Defines the order of exercises in the workout
    sets VARCHAR(50),      -- e.g., '3-5'
    reps VARCHAR(50),      -- e.g., '8-12'
    notes TEXT,
    PRIMARY KEY (workout_id, exercise_id)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
