DROP TRIGGER IF EXISTS habits_updated_at_trigger ON habits;

DROP FUNCTION IF EXISTS update_habits_updated_at();

DROP INDEX IF EXISTS idx_habits_user_active;
DROP INDEX IF EXISTS idx_habits_timezone_deadline;
DROP INDEX IF EXISTS idx_habits_next_deadline;
DROP INDEX IF EXISTS idx_habits_user_id;

DROP TABLE IF EXISTS habits;

DROP TYPE IF EXISTS schedule_type;
