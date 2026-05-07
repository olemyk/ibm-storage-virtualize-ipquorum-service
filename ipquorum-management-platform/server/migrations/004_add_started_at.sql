-- Add started_at column to instances table
ALTER TABLE instances ADD COLUMN started_at TIMESTAMP NULL;

-- Add index for better query performance
CREATE INDEX idx_instances_started_at ON instances(started_at);

-- Made with Bob
