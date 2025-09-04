package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"password-manager/internal/common/crypto"
	"password-manager/internal/common/protocol"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreateUser(username, password string) error {
	hash, salt, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		"INSERT INTO users (username, password_hash, password_salt) VALUES ($1, $2, $3)",
		username, hash, salt,
	)
	return err
}

func (d *Database) AuthenticateUser(username, password string) (bool, error) {
	var hash, salt string
	err := d.db.QueryRow(
		"SELECT password_hash, password_salt FROM users WHERE username = $1",
		username,
	).Scan(&hash, &salt)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return crypto.VerifyPassword(password, hash, salt), nil
}

func (d *Database) GetUserID(username string) (int, error) {
	var userID int
	err := d.db.QueryRow(
		"SELECT id FROM users WHERE username = $1",
		username,
	).Scan(&userID)

	return userID, err
}

func (d *Database) StoreData(userID int, item protocol.NewDataItem) error {
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}

	log.Printf("Storing data for user %d: type=%d, name=%s, data_len=%d", userID, item.Type, item.Name, len(item.Data))

	result, err := d.db.Exec(`
		INSERT INTO user_data (user_id, data_type, name, data, metadata)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, item.Type, item.Name, item.Data, metadataJSON)

	if err != nil {
		log.Printf("SQL error: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
	} else {
		log.Printf("Rows affected: %d", rowsAffected)
	}

	return nil
}

func (d *Database) GetData(userID int, lastSync time.Time) ([]protocol.DataItem, error) {
	rows, err := d.db.Query(`
		SELECT id, data_type, name, data, metadata, created_at, updated_at
		FROM user_data
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at
	`, userID, lastSync)

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

func (d *Database) GetDataByID(userID int, itemID string) (protocol.DataItem, error) {
	var item protocol.DataItem
	var metadataJSON []byte

	err := d.db.QueryRow(`
		SELECT id, data_type, name, data, metadata, created_at, updated_at
		FROM user_data
		WHERE user_id = $1 AND id = $2
	`, userID, itemID).Scan(
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

func (d *Database) DeleteData(userID int, itemID string) error {
	_, err := d.db.Exec(
		"DELETE FROM user_data WHERE user_id = $1 AND id = $2",
		userID, itemID,
	)
	return err
}

func (d *Database) RunMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			password_salt TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS user_data (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			data_type SMALLINT NOT NULL,
			name VARCHAR(255) NOT NULL,
			data BYTEA NOT NULL,
			metadata JSONB DEFAULT '{}'::jsonb,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_user_data_user_id ON user_data(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_data_updated_at ON user_data(updated_at)`,
	}

	for i, migration := range migrations {
		_, err := d.db.Exec(migration)
		if err != nil {
			return fmt.Errorf("migration %d failed: %v", i+1, err)
		}
	}

	return nil
}

func (d *Database) UpdateData(userID int, itemID string, item protocol.NewDataItem) error {
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}

	log.Printf("Updating data for user %d, item %s: type=%d, name=%s, data_len=%d",
		userID, itemID, item.Type, item.Name, len(item.Data))

	_, err = d.db.Exec(`
		UPDATE user_data 
		SET data_type = $1, name = $2, data = $3, metadata = $4, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $5 AND id = $6
	`, item.Type, item.Name, item.Data, metadataJSON, userID, itemID)

	return err
}
