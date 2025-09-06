// Package server предоставляет серверную реализацию менеджера паролей.
//
// Сервер обеспечивает:
// - Многопоточную обработку клиентских соединений
// - Взаимодействие с базой данных PostgreSQL
// - Управление миграциями базы данных
// - Обработку всех операций протокола
// - Аутентификацию и авторизацию пользователей
package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

// Server представляет основной сервер приложения.
// Управляет сетевыми соединениями и взаимодействием с базой данных.
type Server struct {
	host     string
	port     int
	database *Database
}

// NewServer создает новый экземпляр сервера.
//
// Parameters:
//
//	host - хост для прослушивания
//	port - порт для прослушивания
//	dbConnStr - строка подключения к PostgreSQL
//
// Returns:
//
//	*Server - новый экземпляр сервера
//	error - ошибка инициализации
//
// Example:
//
//	connStr := "host=localhost user=postgres dbname=password_manager sslmode=disable"
//	srv, err := NewServer("localhost", 8080, connStr)
func NewServer(host string, port int, dbConnStr string) (*Server, error) {
	db, err := NewDatabase(dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := db.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return &Server{
		host:     host,
		port:     port,
		database: db,
	}, nil
}

// Start запускает сервер и начинает прослушивание подключений.
//
// Returns:
//
//	error - ошибка запуска сервера
func (s *Server) Start() error {
	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server started on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		handler := NewClientHandler(conn, s.database)
		go handler.Handle()
	}
}

// Stop останавливает сервер и освобождает ресурсы.
//
// Returns:
//
//	error - ошибка остановки
func (s *Server) Stop() error {
	return s.database.Close()
}
