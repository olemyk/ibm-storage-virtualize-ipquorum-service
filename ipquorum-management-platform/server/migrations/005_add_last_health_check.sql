-- Add last_health_check column to instances table
ALTER TABLE instances ADD COLUMN last_health_check TIMESTAMP NULL;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_instances_last_health_check ON instances(last_health_check);

-- Made with Bob
