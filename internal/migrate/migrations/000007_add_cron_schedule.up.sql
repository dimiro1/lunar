-- Add cron scheduling columns to functions table
ALTER TABLE functions ADD COLUMN cron_schedule TEXT;
ALTER TABLE functions ADD COLUMN cron_status TEXT DEFAULT 'paused';
