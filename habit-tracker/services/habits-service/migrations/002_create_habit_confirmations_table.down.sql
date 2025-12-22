-- Drop indexes
DROP INDEX IF EXISTS idx_confirmations_habit_date;
DROP INDEX IF EXISTS idx_confirmations_date;
DROP INDEX IF EXISTS idx_confirmations_user_id;
DROP INDEX IF EXISTS idx_confirmations_habit_id;

-- Drop table
DROP TABLE IF EXISTS habit_confirmations;
