package gq

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// **************************************************
// GQ - Gorm Queries
// GQ is a wrapper around gorm generic sql queries.
// **************************************************

// Constants for validation
const (
	MaxPageSize     = 1000
	MaxBatchSize    = 1000
	MaxFieldLength  = 100
	MaxOrderByItems = 5
)

// Errors for validation
var (
	ErrInvalidFieldName  = errors.New("invalid field name")
	ErrInvalidOrderBy    = errors.New("invalid order by clause")
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrInvalidBatchSize  = errors.New("invalid batch size")
	ErrEmptyFilterValue  = errors.New("empty filter value")
	ErrFieldNotFound     = errors.New("field not found")
)

// fieldNameRegex validates field names to prevent SQL injection
var fieldNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// orderByRegex validates ORDER BY clauses
var orderByRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\s+(ASC|DESC|asc|desc))?(\s*,\s*[a-zA-Z][a-zA-Z0-9_]*(\s+(ASC|DESC|asc|desc))?)*$`)

// validateFieldName checks if a field name is safe for SQL queries
func validateFieldName(fieldName string) error {
	if fieldName == "" {
		return fmt.Errorf("%w: field name cannot be empty", ErrInvalidFieldName)
	}

	if len(fieldName) > MaxFieldLength {
		return fmt.Errorf("%w: field name too long (max %d characters)", ErrInvalidFieldName, MaxFieldLength)
	}

	if !fieldNameRegex.MatchString(fieldName) {
		return fmt.Errorf("%w: field name contains invalid characters", ErrInvalidFieldName)
	}

	return nil
}

// validateOrderBy checks if an ORDER BY clause is safe
func validateOrderBy(orderBy string) error {
	if orderBy == "" {
		return nil // Empty is allowed
	}

	if len(orderBy) > MaxFieldLength*MaxOrderByItems {
		return fmt.Errorf("%w: order by clause too long", ErrInvalidOrderBy)
	}

	if !orderByRegex.MatchString(strings.TrimSpace(orderBy)) {
		return fmt.Errorf("%w: invalid order by syntax", ErrInvalidOrderBy)
	}

	return nil
}

// validatePagination checks pagination parameters
func validatePagination(page, pageSize int) error {
	if page < 1 {
		return fmt.Errorf("%w: page must be >= 1", ErrInvalidPagination)
	}

	if pageSize < 1 || pageSize > MaxPageSize {
		return fmt.Errorf("%w: page size must be between 1 and %d", ErrInvalidPagination, MaxPageSize)
	}

	return nil
}

// validateBatchSize checks batch operation size
func validateBatchSize(batchSize int) error {
	if batchSize < 1 || batchSize > MaxBatchSize {
		return fmt.Errorf("%w: batch size must be between 1 and %d", ErrInvalidBatchSize, MaxBatchSize)
	}

	return nil
}

// isFieldInModel checks if a field exists in the given model type using reflection
func isFieldInModel[T any](fieldName string) bool {
	var model T
	modelType := reflect.TypeOf(model)

	// Handle pointer types
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		return false
	}

	// Check all struct fields
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Check gorm column tag first
		gormTag := field.Tag.Get("gorm")
		if gormTag != "" {
			// Parse gorm tag for column name
			parts := strings.Split(gormTag, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "column:") {
					columnName := strings.TrimPrefix(part, "column:")
					if columnName == fieldName {
						return true
					}
				}
			}
		}

		// Check json tag for field mapping
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			// Remove options like ",omitempty"
			jsonName := strings.Split(jsonTag, ",")[0]
			if jsonName == fieldName {
				return true
			}
		}

		// Check if the field name matches directly (PascalCase to snake_case conversion)
		if strings.EqualFold(field.Name, fieldName) {
			return true
		}

		// Check snake_case conversion
		if pascalToSnakeCase(field.Name) == fieldName {
			return true
		}
	}

	return false
}

// snakeToPascalCase converts snake_case to PascalCase
func snakeToPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// pascalToSnakeCase converts PascalCase to snake_case
func pascalToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// validateFilterValue checks filter values for special cases
func validateFilterValue(field string, value interface{}) error {
	if value == nil {
		return nil
	}

	// Special validation for price and pct_remaining fields
	if field == "price" || field == "pct_remaining" {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value for %s must be a string", field)
		}

		if len(str) == 0 {
			return fmt.Errorf("%w: %s value cannot be empty", ErrEmptyFilterValue, field)
		}

		// Check if it has the expected format (number followed by + or -)
		if len(str) < 2 {
			return fmt.Errorf("invalid %s format: must be number followed by + or -", field)
		}
	}

	return nil
}

// InsertRecord inserts a record into the database.
func InsertRecord[T any](db *gorm.DB, record T) (*T, error) {
	result := db.Create(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// BatchInsert inserts a batch of records into the database.
func BatchInsert[T any](db *gorm.DB, records []T, batchSize int) error {
	if err := validateBatchSize(batchSize); err != nil {
		return err
	}

	if len(records) == 0 {
		return nil // Nothing to insert
	}

	if err := db.CreateInBatches(records, batchSize).Error; err != nil {
		return err
	}
	return nil
}

// GetAllRecords gets all records from the database.
func GetAllRecords[T any](db *gorm.DB, page, pageSize int) ([]T, int, error) {
	if err := validatePagination(page, pageSize); err != nil {
		return nil, 0, err
	}

	var records []T
	var totalRecords int64

	// Count total number of records in the table
	if err := db.Model(new(T)).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := db.Offset(offset).Limit(pageSize).Find(&records)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	// Calculate the total number of pages
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))

	return records, totalPages, nil
}

// GetRecordByID gets a record from the database by ID.
func GetRecordByID[T any](db *gorm.DB, id string) (*T, error) {
	var record T
	result := db.Where("id = ?", id).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// GetRecordByField gets a record from the database by field.
func GetRecordByField[T any](db *gorm.DB, fieldName string, fieldValue interface{}) (*T, error) {
	if err := validateFieldName(fieldName); err != nil {
		return nil, err
	}

	if !isFieldInModel[T](fieldName) {
		return nil, fmt.Errorf("%w: field '%s' not found in model", ErrFieldNotFound, fieldName)
	}

	var record T

	result := db.Where(fmt.Sprintf("%s = ?", fieldName), fieldValue).First(&record)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, nil if no record is found
		}
		return nil, result.Error // Return the error for other cases
	}

	return &record, nil
}

// LockAndGetRecordByField gets a record from the database by field and locks the record.
func LockAndGetRecordByField[T any](db *gorm.DB, field string, value interface{}) (*T, error) {
	var record T
	result := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where(fmt.Sprintf("%s = ?", field), value).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// GetRecordsByField gets records from the database by field.
func GetRecordsByField[T any](db *gorm.DB, field string, value interface{}, page, pageSize int, orderBy string) ([]T, int64, error) {
	if err := validateFieldName(field); err != nil {
		return nil, 0, err
	}

	if err := validatePagination(page, pageSize); err != nil {
		return nil, 0, err
	}

	if err := validateOrderBy(orderBy); err != nil {
		return nil, 0, err
	}

	if !isFieldInModel[T](field) {
		return nil, 0, fmt.Errorf("%w: field '%s' not found in model", ErrFieldNotFound, field)
	}

	var records []T
	var totalCount int64

	// Count total records
	countQuery := db.Model(new(T)).Where(fmt.Sprintf("%s = ?", field), value)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Prepare query
	query := db.Where(fmt.Sprintf("%s = ?", field), value)
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// Execute query
	result := query.Find(&records)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return records, totalCount, nil
}

// GetRecordsByFields gets records from the database by fields.
func GetRecordsByFields[T any](db *gorm.DB, conditions map[string]interface{}) ([]T, error) {
	if len(conditions) == 0 {
		return nil, fmt.Errorf("conditions cannot be empty")
	}

	// Validate all field names first
	for field := range conditions {
		if err := validateFieldName(field); err != nil {
			return nil, fmt.Errorf("invalid field '%s': %w", field, err)
		}

		if !isFieldInModel[T](field) {
			return nil, fmt.Errorf("%w: field '%s' not found in model", ErrFieldNotFound, field)
		}
	}

	var records []T

	query := db
	for field, value := range conditions {
		query = query.Where(field+" = ?", value)
	}

	result := query.Find(&records)

	if result.Error != nil {
		return nil, result.Error
	}

	return records, nil
}

// GetFilteredPaginatedRecords gets filtered paginated records from the database.
func GetFilteredPaginatedRecords[T any](db *gorm.DB, page, pageSize int, conditions map[string]interface{}) ([]T, int, error) {
	if err := validatePagination(page, pageSize); err != nil {
		return nil, 0, err
	}

	if len(conditions) == 0 {
		return nil, 0, fmt.Errorf("conditions cannot be empty")
	}

	// Validate all field names and values first
	for field, value := range conditions {
		if err := validateFieldName(field); err != nil {
			return nil, 0, fmt.Errorf("invalid field '%s': %w", field, err)
		}

		if !isFieldInModel[T](field) {
			return nil, 0, fmt.Errorf("%w: field '%s' not found in model", ErrFieldNotFound, field)
		}

		if err := validateFilterValue(field, value); err != nil {
			return nil, 0, fmt.Errorf("invalid value for field '%s': %w", field, err)
		}
	}

	var records []T
	var totalRecords int64

	query := db.Model(new(T)) // Apply model to the query for proper counting

	// Apply conditions dynamically
	for field, value := range conditions {
		if field == "price" {
			str := value.(string) // Already validated above

			// Safe to access since we validated length above
			lastChar := str[len(str)-1]
			priceStr := str[:len(str)-1] // Remove last character for conversion

			price, err := strconv.Atoi(priceStr)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid price value: %v", err)
			}

			if string(lastChar) == "-" {
				query = query.Where(field+" <= ?", price)
				continue
			}

			if string(lastChar) == "+" {
				query = query.Where(field+" >= ?", price)
				continue
			}
		}

		if field == "pct_remaining" {
			str := value.(string) // Already validated above

			// Safe to access since we validated length above
			lastChar := str[len(str)-1]
			pctStr := str[:len(str)-1] // Remove last character for conversion

			pctRemaining, err := strconv.ParseFloat(pctStr, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid pct_remaining value: %v", err)
			}

			if string(lastChar) == "-" {
				query = query.Where(field+" <= ?", pctRemaining)
				continue
			}

			if string(lastChar) == "+" {
				query = query.Where(field+" >= ?", pctRemaining)
				continue
			}
		}

		// Default condition for equality
		query = query.Where(field+" = ?", value)
	}

	// Count total number of records after applying conditions
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Calculate total pages
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))

	// Apply pagination
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Find(&records)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return records, totalPages, nil
}

// UpdateRecordByID updates a record in the database by ID.
func UpdateRecordByID[T any, U any](db *gorm.DB, id string, updates U) error {
	var record T
	result := db.Model(&record).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// LockAndUpdateRecordByID updates a record in the database by ID and locks the record.
func LockAndUpdateRecordByID[T any, U any](db *gorm.DB, id string, updates U) error {
	var record T
	result := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Updates(updates).First(&record)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateRecordByField updates a record in the database by field.
func UpdateRecordByField[T any, U any](db *gorm.DB, field string, value interface{}, updates U) error {
	if err := validateFieldName(field); err != nil {
		return err
	}

	if !isFieldInModel[T](field) {
		return fmt.Errorf("%w: field '%s' not found in model", ErrFieldNotFound, field)
	}

	var record T
	result := db.Model(&record).Where(field+" = ?", value).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteRecordByID deletes a record in the database by ID.
func DeleteRecordByID[T any](db *gorm.DB, id string) error {
	var record T
	result := db.Where("id = ?", id).Delete(&record)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// StringMap is a custom type for handling map[string]string in GORM
type StringMap map[string]string

// Value returns a value for a StringMap
func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan scans a value into a StringMap
func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for StringMap")
	}

	return json.Unmarshal(bytes, m)
}

// InterfaceMap is a custom type for handling map[string]interface{} in GORM
type InterfaceMap map[string]interface{}

// Value returns a value for an InterfaceMap
func (m InterfaceMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan scans a value into an InterfaceMap
func (m *InterfaceMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for InterfaceMap")
	}

	return json.Unmarshal(bytes, m)
}

// **************************************************
// --------------------------------------------------
// GORM Connection & Auto-Migration Utilities
// --------------------------------------------------
// **************************************************

// GormConfig represents GORM configuration
type GormConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
	ConnMaxIdleTime int // in minutes
	LogLevel        string
	SlowThreshold   int // in milliseconds
}

// GormConnection represents a GORM connection wrapper
type GormConnection struct {
	DB     *gorm.DB
	Config *GormConfig
}

// NewGormConnection creates a new GORM connection
func NewGormConnection(config *GormConfig) (*GormConnection, error) {
	// Open database connection
	db, err := gorm.Open(getDialector(config.Driver, config.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Minute)
	}
	if config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Minute)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &GormConnection{
		DB:     db,
		Config: config,
	}, nil
}

// AutoMigrate performs auto-migration for the given models
func (gc *GormConnection) AutoMigrate(models ...interface{}) error {
	if err := gc.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}
	return nil
}

// AutoMigrateWithOptions performs auto-migration with custom options
func (gc *GormConnection) AutoMigrateWithOptions(models []interface{}) error {
	if err := gc.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migration with options failed: %w", err)
	}
	return nil
}

// MigrateTable creates a table for the given model
func (gc *GormConnection) MigrateTable(model interface{}) error {
	if err := gc.DB.Migrator().CreateTable(model); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// DropTable drops a table for the given model
func (gc *GormConnection) DropTable(model interface{}) error {
	if err := gc.DB.Migrator().DropTable(model); err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}
	return nil
}

// HasTable checks if a table exists
func (gc *GormConnection) HasTable(model interface{}) bool {
	return gc.DB.Migrator().HasTable(model)
}

// AddColumn adds a column to a table
func (gc *GormConnection) AddColumn(model interface{}, field string) error {
	if err := gc.DB.Migrator().AddColumn(model, field); err != nil {
		return fmt.Errorf("failed to add column %s: %w", field, err)
	}
	return nil
}

// DropColumn drops a column from a table
func (gc *GormConnection) DropColumn(model interface{}, field string) error {
	if err := gc.DB.Migrator().DropColumn(model, field); err != nil {
		return fmt.Errorf("failed to drop column %s: %w", field, err)
	}
	return nil
}

// HasColumn checks if a column exists
func (gc *GormConnection) HasColumn(model interface{}, field string) bool {
	return gc.DB.Migrator().HasColumn(model, field)
}

// CreateIndex creates an index
func (gc *GormConnection) CreateIndex(model interface{}, name string) error {
	if err := gc.DB.Migrator().CreateIndex(model, name); err != nil {
		return fmt.Errorf("failed to create index %s: %w", name, err)
	}
	return nil
}

// DropIndex drops an index
func (gc *GormConnection) DropIndex(model interface{}, name string) error {
	if err := gc.DB.Migrator().DropIndex(model, name); err != nil {
		return fmt.Errorf("failed to drop index %s: %w", name, err)
	}
	return nil
}

// HasIndex checks if an index exists
func (gc *GormConnection) HasIndex(model interface{}, name string) bool {
	return gc.DB.Migrator().HasIndex(model, name)
}

// Close closes the database connection
func (gc *GormConnection) Close() error {
	sqlDB, err := gc.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Ping tests the database connection
func (gc *GormConnection) Ping() error {
	sqlDB, err := gc.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// Stats returns database connection statistics
func (gc *GormConnection) Stats() (map[string]interface{}, error) {
	sqlDB, err := gc.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
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
	}, nil
}

// Transaction executes a function within a transaction
func (gc *GormConnection) Transaction(fn func(*gorm.DB) error) error {
	return gc.DB.Transaction(fn)
}

// Begin begins a new transaction
func (gc *GormConnection) Begin() *gorm.DB {
	return gc.DB.Begin()
}

// WithContext returns a DB instance with context
func (gc *GormConnection) WithContext(ctx context.Context) *gorm.DB {
	return gc.DB.WithContext(ctx)
}

// Raw executes raw SQL
func (gc *GormConnection) Raw(sql string, values ...interface{}) *gorm.DB {
	return gc.DB.Raw(sql, values...)
}

// Exec executes raw SQL without returning rows
func (gc *GormConnection) Exec(sql string, values ...interface{}) *gorm.DB {
	return gc.DB.Exec(sql, values...)
}

// **************************************************
// --------------------------------------------------
// Migration Utilities
// --------------------------------------------------
// **************************************************

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// Migrator manages database migrations
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator creates a new migrator
func NewMigrator(db *gorm.DB) *Migrator {
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

	if err := m.db.Exec(query).Error; err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// RunMigrations runs all pending migrations
func (m *Migrator) RunMigrations() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	for _, migration := range m.migrations {
		// Check if migration already applied
		var count int64
		err := m.db.Raw("SELECT COUNT(*) FROM migrations WHERE version = ?", migration.Version).Scan(&count).Error
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
		if err := m.db.Exec("INSERT INTO migrations (version, description) VALUES (?, ?)",
			migration.Version, migration.Description).Error; err != nil {
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
		LIMIT ?
	`

	var versions []string
	if err := m.db.Raw(query, count).Scan(&versions).Error; err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
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
		if err := m.db.Exec("DELETE FROM migrations WHERE version = ?", version).Error; err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", version, err)
		}
	}

	return nil
}

// GetAppliedMigrations returns list of applied migrations
func (m *Migrator) GetAppliedMigrations() ([]string, error) {
	var versions []string
	if err := m.db.Raw("SELECT version FROM migrations ORDER BY applied_at ASC").Scan(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	return versions, nil
}

// **************************************************
// --------------------------------------------------
// Helper Functions
// --------------------------------------------------
// **************************************************

// getDialector returns the appropriate GORM dialector based on driver
func getDialector(driver, dsn string) gorm.Dialector {
	switch strings.ToLower(driver) {
	case "postgres", "postgresql":
		return postgres.Open(dsn)
	case "mysql":
		return mysql.Open(dsn)
	case "sqlite", "sqlite3":
		return sqlite.Open(dsn)
	default:
		// Default to postgres
		return postgres.Open(dsn)
	}
}
