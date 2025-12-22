CREATE TYPE schedule_type AS ENUM ('interval', 'weekly');

CREATE TABLE IF NOT EXISTS habits (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    
    -- Basic info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7), -- HEX color, e.g., "#FF5722"
    
    -- Schedule configuration
    schedule_type schedule_type NOT NULL,
    interval_days INTEGER CHECK (interval_days > 0), -- Required for 'interval' type
    weekly_days INTEGER[] CHECK (array_length(weekly_days, 1) > 0), -- Required for 'weekly' type, values 0-6

    -- Timezone offset in hours from UTC (e.g., +3 for Moscow, -5 for New York EST, -8 for Los Angeles PST)
    timezone_offset_hours INTEGER NOT NULL DEFAULT 0 CHECK (timezone_offset_hours >= -12 AND timezone_offset_hours <= 14),
    
    -- Streak state
    streak INTEGER NOT NULL DEFAULT 0 CHECK (streak >= 0),
    next_deadline_utc TIMESTAMP NOT NULL,
    confirmed_for_current_period BOOLEAN NOT NULL DEFAULT FALSE,
    last_confirmed_at TIMESTAMP,
    
    -- Metadata
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_schedule_interval CHECK (
        (schedule_type = 'interval' AND interval_days IS NOT NULL AND weekly_days IS NULL) OR
        (schedule_type = 'weekly' AND weekly_days IS NOT NULL AND interval_days IS NULL)
    ),
    CONSTRAINT valid_weekly_days CHECK (
        weekly_days IS NULL OR array_length(weekly_days, 1) <= 7
    )
);

-- Create indexes for performance
CREATE INDEX idx_habits_user_id ON habits(user_id) WHERE is_active = TRUE;
CREATE INDEX idx_habits_next_deadline ON habits(next_deadline_utc) WHERE is_active = TRUE AND confirmed_for_current_period = FALSE;
CREATE INDEX idx_habits_timezone_deadline ON habits(timezone_offset_hours, next_deadline_utc) WHERE is_active = TRUE AND confirmed_for_current_period = FALSE;
CREATE INDEX idx_habits_user_active ON habits(user_id, is_active);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_habits_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for updated_at
CREATE TRIGGER habits_updated_at_trigger
    BEFORE UPDATE ON habits
    FOR EACH ROW
    EXECUTE FUNCTION update_habits_updated_at();

