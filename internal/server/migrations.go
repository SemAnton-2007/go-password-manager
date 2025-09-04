package server

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type MigrationManager struct {
	db            *sql.DB
	migrationsDir string
}

func NewMigrationManager(db *sql.DB, migrationsDir string) *MigrationManager {
	return &MigrationManager{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

func (m *MigrationManager) RunMigrations() error {
	// Создаем таблицу миграций если она не существует
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Получаем список уже примененных миграций
	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	// Получаем список всех доступных миграций
	availableMigrations, err := m.getAvailableMigrations()
	if err != nil {
		return fmt.Errorf("failed to get available migrations: %v", err)
	}

	// Применяем миграции которые еще не были применены
	for _, migration := range availableMigrations {
		if _, exists := appliedMigrations[migration]; !exists {
			if err := m.applyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %v", migration, err)
			}
			log.Printf("Applied migration: %s", migration)
		}
	}

	return nil
}

func (m *MigrationManager) createMigrationsTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (m *MigrationManager) getAppliedMigrations() (map[string]bool, error) {
	rows, err := m.db.Query("SELECT name FROM migrations ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations[name] = true
	}

	return migrations, nil
}

func (m *MigrationManager) getAvailableMigrations() ([]string, error) {
	files, err := ioutil.ReadDir(m.migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("migrations directory '%s' does not exist", m.migrationsDir)
		}
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}

	// Сортируем миграции по имени
	sort.Strings(migrations)
	return migrations, nil
}

func (m *MigrationManager) applyMigration(migrationName string) error {
	filePath := filepath.Join(m.migrationsDir, migrationName)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %v", migrationName, err)
	}

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Migration %s: some objects already exist, continuing", migrationName)
		} else {
			return fmt.Errorf("failed to execute migration %s: %v", migrationName, err)
		}
	}

	_, err = tx.Exec("INSERT INTO migrations (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", migrationName)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			log.Printf("Migration %s already recorded", migrationName)
		} else {
			return err
		}
	}

	return tx.Commit()
}
