// Package migration предоставляет систему миграций базы данных для менеджера паролей.
//
// Использует библиотеку github.com/golang-migrate/migrate/v4 для управления миграциями.
// Миграции хранятся в виде SQL-файлов в директории migrations.
package server

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// MigrationManager управляет применением миграций базы данных.
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
func (m *MigrationManager) RunMigrations() error {
	absPath, err := filepath.Abs(m.migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path to migrations: %w", err)
	}

	config := m.db.Config()
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		config.ConnConfig.Host,
		config.ConnConfig.Port,
		config.ConnConfig.Database,
		config.ConnConfig.User,
		config.ConnConfig.Password,
	)

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to create sql.DB connection: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	log.Println("Applying database migrations...")
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply")
	} else {
		log.Println("Migrations applied successfully")
	}

	return nil
}
