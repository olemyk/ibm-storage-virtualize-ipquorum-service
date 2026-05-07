package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Database represents the database connection
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	migrations := []string{
		// Instances table
		`CREATE TABLE IF NOT EXISTS instances (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			server_id TEXT DEFAULT 'local',
			api_endpoint TEXT NOT NULL,
			username TEXT NOT NULL,
			storage_system TEXT,
			description TEXT,
			location TEXT,
			status TEXT NOT NULL DEFAULT 'stopped',
			health TEXT NOT NULL DEFAULT 'unknown',
			enable_download BOOLEAN DEFAULT 1,
			enable_mkquorumapp BOOLEAN DEFAULT 1,
			partnersystem TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Servers table
		`CREATE TABLE IF NOT EXISTS servers (
			id TEXT PRIMARY KEY,
			hostname TEXT UNIQUE NOT NULL,
			ip_address TEXT NOT NULL,
			agent_port INTEGER DEFAULT 8444,
			status TEXT NOT NULL DEFAULT 'offline',
			version TEXT,
			last_seen DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'viewer',
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,

		// Health checks table
		`CREATE TABLE IF NOT EXISTS health_checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			instance_id TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL,
			health TEXT NOT NULL,
			network_reachable BOOLEAN,
			network_message TEXT,
			response_time_ms INTEGER,
			checked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			error_message TEXT,
			FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
		)`,

		// Metrics table
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			instance_id TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			cpu_usage REAL,
			memory_usage_bytes INTEGER,
			network_rx_bytes INTEGER,
			network_tx_bytes INTEGER,
			uptime_seconds INTEGER,
			FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
		)`,

		// Audit log table
		`CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT,
			action TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id TEXT,
			details TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		// Schema migrations for existing tables
		// Add email column to users table if it doesn't exist
		`ALTER TABLE users ADD COLUMN email TEXT DEFAULT 'user@ipquorum.local'`,
		// Add active column to users table if it doesn't exist
		`ALTER TABLE users ADD COLUMN active BOOLEAN DEFAULT 1`,

		// Add new agent enhanced fields to instances table
		`ALTER TABLE instances ADD COLUMN ipquorum_name TEXT`,
		`ALTER TABLE instances ADD COLUMN ip6 BOOLEAN DEFAULT 0`,
		`ALTER TABLE instances ADD COLUMN partnerip6 BOOLEAN DEFAULT 0`,
		`ALTER TABLE instances ADD COLUMN nometadata BOOLEAN DEFAULT 0`,

		// Indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_instances_name ON instances(name)`,
		`CREATE INDEX IF NOT EXISTS idx_instances_server_id ON instances(server_id)`,
		`CREATE INDEX IF NOT EXISTS idx_health_checks_instance_id ON health_checks(instance_id)`,
		`CREATE INDEX IF NOT EXISTS idx_health_checks_timestamp ON health_checks(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_instance_id ON metrics(instance_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp)`,
	}

	for _, migration := range migrations {
		if _, err := d.db.Exec(migration); err != nil {
			// Ignore "duplicate column" errors from ALTER TABLE
			if !strings.Contains(err.Error(), "duplicate column") {
				return fmt.Errorf("migration failed: %w", err)
			}
		}
	}

	return nil
}

// CreateDefaultAdmin creates the default admin user
func (d *Database) CreateDefaultAdmin() error {
	// Check if admin already exists
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "admin").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing admin: %w", err)
	}

	if count > 0 {
		return nil // Admin already exists
	}

	// Hash default password
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	id := uuid.New().String()
	_, err = d.db.Exec(
		"INSERT INTO users (id, username, email, password_hash, role, active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, "admin", "admin@ipquorum.local", string(hash), "admin", true, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}

// Instance operations

// CreateInstance creates a new instance
func (d *Database) CreateInstance(instance *Instance) error {
	instance.ID = uuid.New().String()
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()
	instance.Status = "stopped"
	instance.Health = "unknown"

	_, err := d.db.Exec(`
		INSERT INTO instances (
			id, name, server_id, api_endpoint, username, storage_system,
			description, location, status, health, enable_download,
			enable_mkquorumapp, partnersystem, ipquorum_name, ip6,
			partnerip6, nometadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		instance.ID, instance.Name, instance.ServerID, instance.APIEndpoint,
		instance.Username, instance.StorageSystem, instance.Description,
		instance.Location, instance.Status, instance.Health,
		instance.EnableDownload, instance.EnableMkquorumapp, instance.Partnersystem,
		instance.IPQuorumName, instance.IP6, instance.PartnerIP6, instance.NoMetadata,
		instance.CreatedAt, instance.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	return nil
}

// GetInstance retrieves an instance by ID
func (d *Database) GetInstance(id string) (*Instance, error) {
	instance := &Instance{}
	err := d.db.QueryRow(`
		SELECT id, name, server_id, api_endpoint, username, storage_system,
			description, location, status, health, started_at, last_health_check,
			enable_download, enable_mkquorumapp, partnersystem, ipquorum_name, ip6,
			partnerip6, nometadata, created_at, updated_at
		FROM instances WHERE id = ?`, id).Scan(
		&instance.ID, &instance.Name, &instance.ServerID, &instance.APIEndpoint,
		&instance.Username, &instance.StorageSystem, &instance.Description,
		&instance.Location, &instance.Status, &instance.Health, &instance.StartedAt,
		&instance.LastHealthCheck, &instance.EnableDownload, &instance.EnableMkquorumapp,
		&instance.Partnersystem, &instance.IPQuorumName, &instance.IP6, &instance.PartnerIP6,
		&instance.NoMetadata, &instance.CreatedAt, &instance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("instance not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Calculate uptime if instance is running and has a start time
	if instance.StartedAt != nil {
		instance.Uptime = int64(time.Since(*instance.StartedAt).Seconds())
	}

	return instance, nil
}

// GetInstanceByName retrieves an instance by name
func (d *Database) GetInstanceByName(name string) (*Instance, error) {
	instance := &Instance{}
	err := d.db.QueryRow(`
		SELECT id, name, server_id, api_endpoint, username, storage_system,
			description, location, status, health, started_at, last_health_check,
			enable_download, enable_mkquorumapp, partnersystem, ipquorum_name, ip6,
			partnerip6, nometadata, created_at, updated_at
		FROM instances WHERE name = ?`, name).Scan(
		&instance.ID, &instance.Name, &instance.ServerID, &instance.APIEndpoint,
		&instance.Username, &instance.StorageSystem, &instance.Description,
		&instance.Location, &instance.Status, &instance.Health, &instance.StartedAt,
		&instance.LastHealthCheck, &instance.EnableDownload, &instance.EnableMkquorumapp,
		&instance.Partnersystem, &instance.IPQuorumName, &instance.IP6, &instance.PartnerIP6,
		&instance.NoMetadata, &instance.CreatedAt, &instance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("instance not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Calculate uptime if instance is running and has a start time
	if instance.StartedAt != nil {
		instance.Uptime = int64(time.Since(*instance.StartedAt).Seconds())
	}

	return instance, nil
}

// ListInstances retrieves all instances
func (d *Database) ListInstances() ([]*Instance, error) {
	rows, err := d.db.Query(`
		SELECT id, name, server_id, api_endpoint, username, storage_system,
			description, location, status, health, started_at, last_health_check,
			enable_download, enable_mkquorumapp, partnersystem, ipquorum_name, ip6,
			partnerip6, nometadata, created_at, updated_at
		FROM instances ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}
	defer rows.Close()

	var instances []*Instance
	for rows.Next() {
		instance := &Instance{}
		err := rows.Scan(
			&instance.ID, &instance.Name, &instance.ServerID, &instance.APIEndpoint,
			&instance.Username, &instance.StorageSystem, &instance.Description,
			&instance.Location, &instance.Status, &instance.Health, &instance.StartedAt,
			&instance.LastHealthCheck, &instance.EnableDownload, &instance.EnableMkquorumapp,
			&instance.Partnersystem, &instance.IPQuorumName, &instance.IP6, &instance.PartnerIP6,
			&instance.NoMetadata, &instance.CreatedAt, &instance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instance: %w", err)
		}

		// Calculate uptime if instance is running and has a start time
		if instance.StartedAt != nil {
			instance.Uptime = int64(time.Since(*instance.StartedAt).Seconds())
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// UpdateInstance updates an instance
func (d *Database) UpdateInstance(instance *Instance) error {
	instance.UpdatedAt = time.Now()

	_, err := d.db.Exec(`
		UPDATE instances SET
			api_endpoint = ?, username = ?, storage_system = ?,
			description = ?, location = ?, status = ?, health = ?,
			enable_download = ?, enable_mkquorumapp = ?, partnersystem = ?,
			ipquorum_name = ?, ip6 = ?, partnerip6 = ?, nometadata = ?,
			started_at = ?, last_health_check = ?, updated_at = ?
		WHERE id = ?`,
		instance.APIEndpoint, instance.Username, instance.StorageSystem,
		instance.Description, instance.Location, instance.Status, instance.Health,
		instance.EnableDownload, instance.EnableMkquorumapp, instance.Partnersystem,
		instance.IPQuorumName, instance.IP6, instance.PartnerIP6, instance.NoMetadata,
		instance.StartedAt, instance.LastHealthCheck, instance.UpdatedAt, instance.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	return nil
}

// DeleteInstance deletes an instance
func (d *Database) DeleteInstance(id string) error {
	_, err := d.db.Exec("DELETE FROM instances WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	return nil
}

// User operations

// ListUsers retrieves all users
func (d *Database) ListUsers() ([]*User, error) {
	rows, err := d.db.Query(`
		SELECT id, username, email, password_hash, role, active, created_at, last_login
		FROM users ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.Active, &user.CreatedAt, &user.LastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUser retrieves a user by ID
func (d *Database) GetUser(id string) (*User, error) {
	user := &User{}
	err := d.db.QueryRow(`
		SELECT id, username, email, password_hash, role, active, created_at, last_login
		FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Active, &user.CreatedAt, &user.LastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (d *Database) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := d.db.QueryRow(`
		SELECT id, username, email, password_hash, role, active, created_at, last_login
		FROM users WHERE username = ?`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.Active, &user.CreatedAt, &user.LastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user
func (d *Database) CreateUser(user *User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()

	_, err := d.db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.Role, user.Active, user.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser updates a user
func (d *Database) UpdateUser(user *User) error {
	_, err := d.db.Exec(`
		UPDATE users SET
			email = ?, password_hash = ?, role = ?, active = ?
		WHERE id = ?`,
		user.Email, user.PasswordHash, user.Role, user.Active, user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (d *Database) DeleteUser(id string) error {
	_, err := d.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// UpdateUserLastLogin updates the last login time for a user
func (d *Database) UpdateUserLastLogin(userID string) error {
	_, err := d.db.Exec("UPDATE users SET last_login = ? WHERE id = ?", time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// Health check operations

// CreateHealthCheck creates a new health check record
func (d *Database) CreateHealthCheck(check *HealthCheck) error {
	check.Timestamp = time.Now()

	result, err := d.db.Exec(`
		INSERT INTO health_checks (
			instance_id, timestamp, status, health, network_reachable,
			network_message, response_time_ms, checked_at, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		check.InstanceID, check.Timestamp, check.Status, check.Health,
		check.NetworkReachable, check.NetworkMessage, check.ResponseTime,
		check.CheckedAt, check.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to create health check: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get health check ID: %w", err)
	}

	check.ID = id
	return nil
}

// GetLatestHealthCheck retrieves the latest health check for an instance
func (d *Database) GetLatestHealthCheck(instanceID string) (*HealthCheck, error) {
	check := &HealthCheck{}
	err := d.db.QueryRow(`
		SELECT id, instance_id, timestamp, status, health, network_reachable,
			network_message, response_time_ms, checked_at, error_message
		FROM health_checks
		WHERE instance_id = ?
		ORDER BY timestamp DESC
		LIMIT 1`, instanceID).Scan(
		&check.ID, &check.InstanceID, &check.Timestamp, &check.Status,
		&check.Health, &check.NetworkReachable, &check.NetworkMessage,
		&check.ResponseTime, &check.CheckedAt, &check.ErrorMessage,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no health checks found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get health check: %w", err)
	}

	return check, nil
}

// Audit log operations

// CreateAuditLog creates a new audit log entry
func (d *Database) CreateAuditLog(log *AuditLog) error {
	log.Timestamp = time.Now()

	result, err := d.db.Exec(`
		INSERT INTO audit_log (
			user_id, action, resource_type, resource_id, details, timestamp
		) VALUES (?, ?, ?, ?, ?, ?)`,
		log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.Details, log.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get audit log ID: %w", err)
	}

	log.ID = id
	return nil
}

// GetHealthChecksByInstance retrieves health check history for an instance
func (d *Database) GetHealthChecksByInstance(instanceID string, limit int) ([]*HealthCheck, error) {
	rows, err := d.db.Query(`
		SELECT id, instance_id, timestamp, status, health, network_reachable,
			network_message, response_time_ms, checked_at, error_message
		FROM health_checks
		WHERE instance_id = ?
		ORDER BY timestamp DESC
		LIMIT ?`, instanceID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query health checks: %w", err)
	}
	defer rows.Close()

	var checks []*HealthCheck
	for rows.Next() {
		check := &HealthCheck{}
		err := rows.Scan(
			&check.ID, &check.InstanceID, &check.Timestamp, &check.Status,
			&check.Health, &check.NetworkReachable, &check.NetworkMessage,
			&check.ResponseTime, &check.CheckedAt, &check.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan health check: %w", err)
		}
		checks = append(checks, check)
	}

	return checks, nil
}

// Ping checks database connectivity
func (d *Database) Ping() error {
	return d.db.Ping()
}

// Server management methods

// CreateServer creates a new server record
func (d *Database) CreateServer(server *Server) error {
	server.ID = uuid.New().String()
	server.CreatedAt = time.Now()

	query := `
		INSERT INTO servers (id, hostname, ip_address, agent_port, status, version, last_seen, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query,
		server.ID,
		server.Hostname,
		server.IPAddress,
		server.AgentPort,
		server.Status,
		server.Version,
		server.LastSeen,
		server.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return nil
}

// GetServer retrieves a server by ID
func (d *Database) GetServer(id string) (*Server, error) {
	query := `
		SELECT id, hostname, ip_address, agent_port, status, version, last_seen, created_at
		FROM servers
		WHERE id = ?
	`

	var server Server
	var lastSeen sql.NullTime
	var version sql.NullString

	err := d.db.QueryRow(query, id).Scan(
		&server.ID,
		&server.Hostname,
		&server.IPAddress,
		&server.AgentPort,
		&server.Status,
		&version,
		&lastSeen,
		&server.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	if version.Valid {
		server.Version = version.String
	}
	if lastSeen.Valid {
		server.LastSeen = &lastSeen.Time
	}

	return &server, nil
}

// ListServers retrieves all servers
func (d *Database) ListServers() ([]*Server, error) {
	query := `
		SELECT id, hostname, ip_address, agent_port, status, version, last_seen, created_at
		FROM servers
		ORDER BY hostname
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	defer rows.Close()

	var servers []*Server
	for rows.Next() {
		var server Server
		var lastSeen sql.NullTime
		var version sql.NullString

		err := rows.Scan(
			&server.ID,
			&server.Hostname,
			&server.IPAddress,
			&server.AgentPort,
			&server.Status,
			&version,
			&lastSeen,
			&server.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}

		if version.Valid {
			server.Version = version.String
		}
		if lastSeen.Valid {
			server.LastSeen = &lastSeen.Time
		}

		servers = append(servers, &server)
	}

	return servers, nil
}

// UpdateServer updates a server record
func (d *Database) UpdateServer(server *Server) error {
	query := `
		UPDATE servers
		SET hostname = ?, ip_address = ?, agent_port = ?, status = ?, version = ?, last_seen = ?
		WHERE id = ?
	`

	_, err := d.db.Exec(query,
		server.Hostname,
		server.IPAddress,
		server.AgentPort,
		server.Status,
		server.Version,
		server.LastSeen,
		server.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	return nil
}

// DeleteServer deletes a server by ID
func (d *Database) DeleteServer(id string) error {
	query := `DELETE FROM servers WHERE id = ?`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("server not found")
	}

	return nil
}

// UpdateServerStatus updates the status and last_seen timestamp of a server
func (d *Database) UpdateServerStatus(id, status string) error {
	now := time.Now()
	query := `
		UPDATE servers
		SET status = ?, last_seen = ?
		WHERE id = ?
	`

	_, err := d.db.Exec(query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update server status: %w", err)
	}

	return nil
}
