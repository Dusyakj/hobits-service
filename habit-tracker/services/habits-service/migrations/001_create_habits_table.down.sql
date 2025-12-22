-- Drop trigger
DROP TRIGGER IF EXISTS habits_updated_at_trigger ON habits;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_habits_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_habits_user_active;
DROP INDEX IF EXISTS idx_habits_timezone_deadline;
DROP INDEX IF EXISTS idx_habits_next_deadline;
DROP INDEX IF EXISTS idx_habits_user_id;

-- Drop table
DROP TABLE IF EXISTS habits;

-- Drop ENUM type
DROP TYPE IF EXISTS schedule_type;
