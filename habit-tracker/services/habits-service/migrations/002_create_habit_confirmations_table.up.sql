CREATE TABLE IF NOT EXISTS habit_confirmations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,

    confirmed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    confirmed_for_date DATE NOT NULL, -- The date this confirmation is for (in habit's timezone)

    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_habit_confirmation UNIQUE (habit_id, confirmed_for_date)
);

CREATE INDEX idx_confirmations_habit_id ON habit_confirmations(habit_id);
CREATE INDEX idx_confirmations_user_id ON habit_confirmations(user_id);
CREATE INDEX idx_confirmations_date ON habit_confirmations(confirmed_for_date DESC);
CREATE INDEX idx_confirmations_habit_date ON habit_confirmations(habit_id, confirmed_for_date DESC);


