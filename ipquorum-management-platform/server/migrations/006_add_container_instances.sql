-- Migration 006: Add container instances support
-- This migration adds support for containerized IPQuorum deployments

-- Add deployment_type column to instances table
ALTER TABLE instances ADD COLUMN deployment_type TEXT NOT NULL DEFAULT 'systemd' 
    CHECK(deployment_type IN ('systemd', 'container'));

-- Create container_instances table
CREATE TABLE IF NOT EXISTS container_instances (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    container_id TEXT NOT NULL,
    container_name TEXT NOT NULL,
    image_name TEXT NOT NULL,
    image_tag TEXT NOT NULL DEFAULT 'latest',
    runtime TEXT NOT NULL CHECK(runtime IN ('docker', 'podman')),
    status TEXT NOT NULL CHECK(status IN ('created', 'running', 'paused', 'stopped', 'error')),
    state TEXT,
    restart_policy TEXT NOT NULL DEFAULT 'no' CHECK(restart_policy IN ('no', 'always', 'on-failure', 'unless-stopped')),
    restart_count INTEGER NOT NULL DEFAULT 0,
    exit_code INTEGER,
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
);

-- Create indexes for container_instances
CREATE INDEX IF NOT EXISTS idx_container_instances_instance_id ON container_instances(instance_id);
CREATE INDEX IF NOT EXISTS idx_container_instances_agent_id ON container_instances(agent_id);
CREATE INDEX IF NOT EXISTS idx_container_instances_status ON container_instances(status);
CREATE INDEX IF NOT EXISTS idx_container_instances_runtime ON container_instances(runtime);
CREATE INDEX IF NOT EXISTS idx_container_instances_created_at ON container_instances(created_at DESC);

-- Create container_configs table for storing port mappings, volumes, etc.
CREATE TABLE IF NOT EXISTS container_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    container_instance_id TEXT NOT NULL,
    config_type TEXT NOT NULL CHECK(config_type IN ('port', 'volume', 'env', 'label')),
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    extra TEXT, -- JSON for additional data (e.g., protocol for ports, read_only for volumes)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (container_instance_id) REFERENCES container_instances(id) ON DELETE CASCADE
);

-- Create index for container_configs
CREATE INDEX IF NOT EXISTS idx_container_configs_container_id ON container_configs(container_instance_id);
CREATE INDEX IF NOT EXISTS idx_container_configs_type ON container_configs(config_type);

-- Create container_stats table for historical statistics
CREATE TABLE IF NOT EXISTS container_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    container_instance_id TEXT NOT NULL,
    cpu_percent REAL NOT NULL DEFAULT 0,
    memory_usage INTEGER NOT NULL DEFAULT 0,
    memory_limit INTEGER NOT NULL DEFAULT 0,
    memory_percent REAL NOT NULL DEFAULT 0,
    network_rx_bytes INTEGER NOT NULL DEFAULT 0,
    network_tx_bytes INTEGER NOT NULL DEFAULT 0,
    block_read INTEGER NOT NULL DEFAULT 0,
    block_write INTEGER NOT NULL DEFAULT 0,
    pids INTEGER NOT NULL DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (container_instance_id) REFERENCES container_instances(id) ON DELETE CASCADE
);

-- Create indexes for container_stats
CREATE INDEX IF NOT EXISTS idx_container_stats_container_id ON container_stats(container_instance_id);
CREATE INDEX IF NOT EXISTS idx_container_stats_timestamp ON container_stats(timestamp DESC);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_container_instances_updated_at
    AFTER UPDATE ON container_instances
    FOR EACH ROW
BEGIN
    UPDATE container_instances SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Rollback instructions (commented out, run manually if needed):
-- DROP TRIGGER IF EXISTS update_container_instances_updated_at;
-- DROP TABLE IF EXISTS container_stats;
-- DROP TABLE IF EXISTS container_configs;
-- DROP TABLE IF EXISTS container_instances;
-- ALTER TABLE instances DROP COLUMN deployment_type;

-- Made with Bob
