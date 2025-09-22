// Package migrations provides database migration functionality for the TT Stock Backend API.
// It handles running SQL migrations in order and tracking migration status.
package migrations

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Filename    string
	AppliedAt   *time.Time
}

// MigrationRunner handles running database migrations
type MigrationRunner struct {
	db *sql.DB
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *sql.DB) *MigrationRunner {
	return &MigrationRunner{
		db: db,
	}
}

// RunMigrations executes all pending migrations
func (mr *MigrationRunner) RunMigrations(migrationsDir string) error {
	log.Println("Starting database migrations...")

	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := mr.getMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := mr.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrationFiles {
		if mr.isMigrationApplied(migration.Version, appliedMigrations) {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Running migration: %s - %s", migration.Version, migration.Description)
		if err := mr.runMigration(migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}

		log.Printf("Migration %s completed successfully", migration.Version)
	}

	log.Println("All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func (mr *MigrationRunner) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			description VARCHAR(500),
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := mr.db.Exec(query)
	return err
}

// getMigrationFiles returns a list of migration files sorted by version
func (mr *MigrationRunner) getMigrationFiles(migrationsDir string) ([]Migration, error) {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Parse version and description from filename
		// Format: 001_initial_schema.sql
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")
		description = strings.ReplaceAll(description, "_", " ")

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			Filename:    file.Name(),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a list of applied migrations
func (mr *MigrationRunner) getAppliedMigrations() (map[string]Migration, error) {
	query := `SELECT version, description, applied_at FROM schema_migrations ORDER BY version`
	rows, err := mr.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedMigrations := make(map[string]Migration)
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.Version, &migration.Description, &migration.AppliedAt)
		if err != nil {
			return nil, err
		}
		appliedMigrations[migration.Version] = migration
	}

	return appliedMigrations, nil
}

// isMigrationApplied checks if a migration has already been applied
func (mr *MigrationRunner) isMigrationApplied(version string, appliedMigrations map[string]Migration) bool {
	_, exists := appliedMigrations[version]
	return exists
}

// runMigration executes a single migration
func (mr *MigrationRunner) runMigration(migration Migration) error {
	// Read migration file
	migrationPath := filepath.Join("migrations", migration.Filename)
	content, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Start transaction
	tx, err := mr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	insertQuery := `
		INSERT INTO schema_migrations (version, description, applied_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP)
	`
	if _, err := tx.Exec(insertQuery, migration.Version, migration.Description); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// GetMigrationStatus returns the status of all migrations
func (mr *MigrationRunner) GetMigrationStatus() ([]Migration, error) {
	query := `
		SELECT version, description, applied_at 
		FROM schema_migrations 
		ORDER BY version
	`
	rows, err := mr.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.Version, &migration.Description, &migration.AppliedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// RollbackMigration rolls back a specific migration (use with caution)
func (mr *MigrationRunner) RollbackMigration(version string) error {
	log.Printf("Rolling back migration: %s", version)

	// Start transaction
	tx, err := mr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove migration record
	deleteQuery := `DELETE FROM schema_migrations WHERE version = $1`
	if _, err := tx.Exec(deleteQuery, version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	log.Printf("Migration %s rolled back successfully", version)
	return nil
}
