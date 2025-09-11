// Package migration предоставляет систему миграций базы данных для менеджера паролей.
//
// Миграции позволяют:
// - Создавать и обновлять схему базы данных
// - Управлять версиями схемы
// - Автоматически применять изменения при обновлении
//
// Миграции хранятся в виде SQL-файлов в директории migrations.
package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationManager управляет применением миграций базы данных.
// Отслеживает примененные миграции и обеспечивает их идемпотентность.
type MigrationManager struct {
	db            *pgxpool.Pool
	migrationsDir string
}

// NewMigrationManager создает новый менеджер миграций.
//
// Parameters:
//
//	db            - подключение к базе данных
//	migrationsDir - путь к директории с миграциями
//
// Returns:
//
//	*MigrationManager - новый экземпляр менеджера
func NewMigrationManager(db *pgxpool.Pool, migrationsDir string) *MigrationManager {
	return &MigrationManager{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// RunMigrations применяет все непримененные миграции к базе данных.
//
// Returns:
//
//	error - ошибка применения миграций
//
// Process:
//  1. Создает таблицу миграций если не существует
//  2. Получает список уже примененных миграций
//  3. Находит все доступные миграции в директории
//  4. Применяет миграции в порядке их имен
//  5. Записывает applied миграции в таблицу
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

// createMigrationsTable создает таблицу для отслеживания примененных миграций.
//
// Returns:
//
//	error - ошибка создания таблицы
//
// Table schema:
//
//	CREATE TABLE IF NOT EXISTS migrations (
//	    id SERIAL PRIMARY KEY,
//	    name VARCHAR(255) NOT NULL UNIQUE,
//	    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
//	)
func (m *MigrationManager) createMigrationsTable() error {
	_, err := m.db.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
	)
	return err
}

// getAppliedMigrations возвращает список уже примененных миграций.
//
// Returns:
//
//	map[string]bool - карта имен примененных миграций
//	error - ошибка запроса
func (m *MigrationManager) getAppliedMigrations() (map[string]bool, error) {
	rows, err := m.db.Query(context.Background(), "SELECT name FROM migrations ORDER BY name")
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

// getAvailableMigrations возвращает список всех доступных миграций.
//
// Returns:
//
//	[]string - список имен файлов миграций
//	error - ошибка чтения директории
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

// applyMigration применяет конкретную миграцию к базе данных.
//
// Parameters:
//
//	migrationName - имя файла миграции
//
// Returns:
//
//	error - ошибка применения миграции
func (m *MigrationManager) applyMigration(migrationName string) error {
	filePath := filepath.Join(m.migrationsDir, migrationName)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %v", migrationName, err)
	}

	tx, err := m.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(context.Background(), string(content)); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Migration %s: some objects already exist, continuing", migrationName)
		} else {
			return fmt.Errorf("failed to execute migration %s: %v", migrationName, err)
		}
	}

	_, err = tx.Exec(
		context.Background(),
		"INSERT INTO migrations (name) VALUES ($1) ON CONFLICT (name) DO NOTHING",
		migrationName,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			log.Printf("Migration %s already recorded", migrationName)
		} else {
			return err
		}
	}

	return tx.Commit(context.Background())
}
