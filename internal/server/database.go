// Package database предоставляет реализацию серверной части менеджера паролей.
//
// Включает:
// - Управление подключениями к базе данных PostgreSQL
// - Выполнение миграций базы данных
// - Операции с пользователями и их данными
// - Аутентификацию и авторизацию
package server

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"password-manager/internal/common/crypto"
	"password-manager/internal/common/protocol"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db *pgxpool.Pool
}

// NewDatabase создает новое подключение к базе данных.
//
// Parameters:
//
//	connStr - строка подключения к PostgreSQL
//
// Returns:
//
//	*Database - подключение к базе данных
//	error - ошибка подключения
//
// Example:
//
//	db, err := NewDatabase("host=localhost user=postgres dbname=test")
func NewDatabase(connStr string) (*Database, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &Database{db: pool}, nil
}

// Close закрывает подключение к базе данных.
//
// Returns:
//
//	error - ошибка закрытия соединения
func (d *Database) Close() error {
	d.db.Close()
	return nil
}

// RunMigrations выполняет миграции базы данных.
//
// Returns:
//
//	error - ошибка выполнения миграций
func (d *Database) RunMigrations() error {
	dir, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		return err
	}

	migrationsDir := filepath.Join(dir, "migrations")

	migrationManager := NewMigrationManager(d.db, migrationsDir)
	return migrationManager.RunMigrations()
}

// CreateUser создает нового пользователя в системе.
//
// Parameters:
//
//	username - имя пользователя
//	password - пароль
//
// Returns:
//
//	error - ошибка создания пользователя
func (d *Database) CreateUser(username, password string) error {
	hash, salt, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		context.Background(),
		"INSERT INTO users (username, password_hash, password_salt) VALUES ($1, $2, $3)",
		username, hash, salt,
	)
	return err
}

// AuthenticateUser проверяет credentials пользователя.
//
// Parameters:
//
//	username - имя пользователя
//	password - пароль
//
// Returns:
//
//	bool - true если аутентификация успешна
//	error - ошибка проверки.
func (d *Database) AuthenticateUser(username, password string) (bool, error) {
	var hash, salt string
	err := d.db.QueryRow(
		context.Background(),
		"SELECT password_hash, password_salt FROM users WHERE username = $1",
		username,
	).Scan(&hash, &salt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return crypto.VerifyPassword(password, hash, salt), nil
}

// GetUserID возвращает внутренний ID пользователя по имени.
//
// Parameters:
//
//	username - имя пользователя
//
// Returns:
//
//	int - внутренний ID пользователя
//	error - ошибка если пользователь не найден
func (d *Database) GetUserID(username string) (int, error) {
	var userID int
	err := d.db.QueryRow(
		context.Background(),
		"SELECT id FROM users WHERE username = $1",
		username,
	).Scan(&userID)

	return userID, err
}

// StoreData сохраняет элемент данных для пользователя.
//
// Parameters:
//
//	userID - ID пользователя-владельца
//	item   - элемент данных для сохранения
//
// Returns:
//
//	error - ошибка сохранения
func (d *Database) StoreData(userID int, item protocol.NewDataItem) error {
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}

	log.Printf("Storing data for user %d: type=%d, name=%s, data_len=%d", userID, item.Type, item.Name, len(item.Data))

	_, err = d.db.Exec(
		context.Background(),
		"INSERT INTO user_data (user_id, data_type, name, data, metadata) VALUES ($1, $2, $3, $4, $5)",
		userID, item.Type, item.Name, item.Data, metadataJSON,
	)

	return err
}

// GetData возвращает элементы данных пользователя, измененные после указанного времени.
//
// Parameters:
//
//	userID   - ID пользователя
//	lastSync - время последней синхронизации
//
// Returns:
//
//	[]DataItem - список элементов данных
//	error      - ошибка запроса
func (d *Database) GetData(userID int, lastSync time.Time) ([]protocol.DataItem, error) {
	rows, err := d.db.Query(
		context.Background(),
		`SELECT id, data_type, name, data, metadata, created_at, updated_at
		 FROM user_data
		 WHERE user_id = $1 AND updated_at > $2
		 ORDER BY updated_at`,
		userID, lastSync,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []protocol.DataItem
	for rows.Next() {
		var item protocol.DataItem
		var metadataJSON []byte

		err := rows.Scan(
			&item.ID, &item.Type, &item.Name, &item.Data, &metadataJSON,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &item.Metadata); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

// GetDataByID возвращает конкретный элемент данных по ID.
//
// Parameters:
//
//	userID - ID пользователя-владельца
//	itemID - ID элемента данных
//
// Returns:
//
//	DataItem - найденный элемент данных
//	error    - ошибка если элемент не найден или нет доступа
func (d *Database) GetDataByID(userID int, itemID string) (protocol.DataItem, error) {
	var item protocol.DataItem
	var metadataJSON []byte

	err := d.db.QueryRow(
		context.Background(),
		`SELECT id, data_type, name, data, metadata, created_at, updated_at
		 FROM user_data
		 WHERE user_id = $1 AND id = $2`,
		userID, itemID,
	).Scan(
		&item.ID, &item.Type, &item.Name, &item.Data, &metadataJSON,
		&item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		return protocol.DataItem{}, err
	}

	if err := json.Unmarshal(metadataJSON, &item.Metadata); err != nil {
		return protocol.DataItem{}, err
	}

	return item, nil
}

// DeleteData удаляет элемент данных пользователя.
//
// Parameters:
//
//	userID - ID пользователя-владельца
//	itemID - ID элемента для удаления
//
// Returns:
//
//	error - ошибка удаления
func (d *Database) DeleteData(userID int, itemID string) error {
	_, err := d.db.Exec(
		context.Background(),
		"DELETE FROM user_data WHERE user_id = $1 AND id = $2",
		userID, itemID,
	)
	return err
}

// UpdateData обновляет существующий элемент данных.
//
// Parameters:
//
//	userID - ID пользователя-владельца
//	itemID - ID элемента для обновления
//	item   - новые данные элемента
//
// Returns:
//
//	error - ошибка обновления
func (d *Database) UpdateData(userID int, itemID string, item protocol.NewDataItem) error {
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}

	log.Printf("Updating data for user %d, item %s: type=%d, name=%s, data_len=%d",
		userID, itemID, item.Type, item.Name, len(item.Data))

	_, err = d.db.Exec(
		context.Background(),
		`UPDATE user_data
		 SET data_type = $1, name = $2, data = $3, metadata = $4, updated_at = CURRENT_TIMESTAMP
		 WHERE user_id = $5 AND id = $6`,
		item.Type, item.Name, item.Data, metadataJSON, userID, itemID,
	)

	return err
}
