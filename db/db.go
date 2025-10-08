package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Config represents database configuration
type Config struct {
	Host         string
	Port         int
	Database     string
	Username     string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	MaxIdleTime  time.Duration
}

// Connection represents a database connection
type Connection struct {
	DB     *sql.DB
	Config *Config
}

// NewConnection creates a new database connection
func NewConnection(config *Config) (*Connection, error) {
	// This is a generic interface - specific drivers will implement the connection string
	connStr := buildConnectionString(config)

	db, err := sql.Open("postgres", connStr) // Default to postgres, can be overridden
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxLifetime > 0 {
		db.SetConnMaxLifetime(config.MaxLifetime)
	}
	if config.MaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.MaxIdleTime)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Connection{
		DB:     db,
		Config: config,
	}, nil
}

// buildConnectionString builds a connection string based on config
func buildConnectionString(config *Config) string {
	// This is a simplified version - in production you'd want more robust connection string building
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
}

// Close closes the database connection
func (c *Connection) Close() error {
	return c.DB.Close()
}

// Ping tests the database connection
func (c *Connection) Ping() error {
	return c.DB.Ping()
}

// Stats returns database connection statistics
func (c *Connection) Stats() sql.DBStats {
	return c.DB.Stats()
}

// Transaction represents a database transaction
type Transaction struct {
	tx *sql.Tx
}

// BeginTransaction begins a new transaction
func (c *Connection) BeginTransaction() (*Transaction, error) {
	tx, err := c.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{tx: tx}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// Query executes a query and returns rows within the transaction
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryRow executes a query and returns a single row within the transaction
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// QueryBuilder represents a query builder
type QueryBuilder struct {
	selectFields []string
	fromTable    string
	whereClauses []string
	whereArgs    []interface{}
	orderBy      []string
	limitValue   int
	offsetValue  int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		whereClauses: make([]string, 0),
		whereArgs:    make([]interface{}, 0),
		orderBy:      make([]string, 0),
	}
}

// Select sets the SELECT fields
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selectFields = fields
	return qb
}

// From sets the FROM table
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.fromTable = table
	return qb
}

// Where adds a WHERE clause
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.whereClauses = append(qb.whereClauses, condition)
	qb.whereArgs = append(qb.whereArgs, args...)
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(field string, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", field, direction))
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitValue = limit
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetValue = offset
	return qb
}

// Build builds the final query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := "SELECT "

	// SELECT fields
	if len(qb.selectFields) > 0 {
		for i, field := range qb.selectFields {
			if i > 0 {
				query += ", "
			}
			query += field
		}
	} else {
		query += "*"
	}

	// FROM table
	if qb.fromTable != "" {
		query += " FROM " + qb.fromTable
	}

	// WHERE clauses
	if len(qb.whereClauses) > 0 {
		query += " WHERE "
		for i, clause := range qb.whereClauses {
			if i > 0 {
				query += " AND "
			}
			query += clause
		}
	}

	// ORDER BY
	if len(qb.orderBy) > 0 {
		query += " ORDER BY "
		for i, order := range qb.orderBy {
			if i > 0 {
				query += ", "
			}
			query += order
		}
	}

	// LIMIT
	if qb.limitValue > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limitValue)
	}

	// OFFSET
	if qb.offsetValue > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offsetValue)
	}

	return query, qb.whereArgs
}

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

// Migrator manages database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// AddMigration adds a migration
func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// CreateMigrationsTable creates the migrations table
func (m *Migrator) CreateMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version VARCHAR(255) PRIMARY KEY,
			description TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := m.db.Exec(query)
	return err
}

// RunMigrations runs all pending migrations
func (m *Migrator) RunMigrations() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	for _, migration := range m.migrations {
		// Check if migration already applied
		var count int
		err := m.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = $1", migration.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			continue // Migration already applied
		}

		// Run migration
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}

		// Record migration
		_, err = m.db.Exec("INSERT INTO migrations (version, description) VALUES ($1, $2)",
			migration.Version, migration.Description)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// RollbackMigrations rolls back migrations
func (m *Migrator) RollbackMigrations(count int) error {
	// Get applied migrations in reverse order
	query := `
		SELECT version FROM migrations 
		ORDER BY applied_at DESC 
		LIMIT $1
	`

	rows, err := m.db.Query(query, count)
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		versions = append(versions, version)
	}

	// Rollback migrations
	for _, version := range versions {
		// Find migration
		var migration *Migration
		for _, m := range m.migrations {
			if m.Version == version {
				migration = &m
				break
			}
		}

		if migration == nil {
			return fmt.Errorf("migration %s not found", version)
		}

		// Run down migration
		if err := migration.Down(m.db); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", version, err)
		}

		// Remove migration record
		_, err = m.db.Exec("DELETE FROM migrations WHERE version = $1", version)
		if err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", version, err)
		}
	}

	return nil
}

// HealthCheck performs a database health check
func (c *Connection) HealthCheck() error {
	// Test basic connectivity
	if err := c.DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var result int
	err := c.DB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query returned unexpected result: %d", result)
	}

	return nil
}

// GetConnectionInfo returns connection information
func (c *Connection) GetConnectionInfo() map[string]interface{} {
	stats := c.DB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
