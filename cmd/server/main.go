// Package main предоставляет серверное приложение для менеджера паролей.
//
// Сервер обеспечивает:
// - Аутентификацию и авторизацию пользователей
// - Хранение зашифрованных данных в PostgreSQL
// - Синхронизацию данных между клиентами
// - Обработку сетевых запросов по собственному протоколу
//
// Пример запуска:
//
//	go run cmd/server/main.go -db-host=localhost -db-user=postgres
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"password-manager/internal/server"
)

// main является точкой входа серверного приложения.
//
// Функция выполняет:
//  1. Парсинг аргументов командной строки для конфигурации сервера и БД
//  2. Формирование строки подключения к PostgreSQL
//  3. Создание и инициализацию экземпляра сервера
//  4. Запуск сервера и обработку входящих подключений
//  5. Обработку ошибок и корректное завершение работы
//
// Parameters:
//
//	-db-host     - хост базы данных (по умолчанию: localhost)
//	-db-port     - порт базы данных (по умолчанию: 5432)
//	-db-name     - имя базы данных (по умолчанию: password_manager)
//	-db-user     - пользователь базы данных (по умолчанию: postgres)
//	-db-password - пароль базы данных (обязательный параметр)
//	-db-ssl-mode - режим SSL подключения (по умолчанию: disable)
//	-host        - хост для прослушивания (по умолчанию: localhost)
//	-port        - порт для прослушивания (по умолчанию: 8080)
//
// Exit codes:
//   - 0 - успешное завершение
//   - 1 - ошибка создания или запуска сервера
//
// Example:
//
//	go run cmd/server/main.go -db-host=localhost -db-user=postgres -db-password=secret
func main() {
	host := flag.String("host", "localhost", "Server host")
	port := flag.Int("port", 8080, "Server port")
	dbHost := flag.String("db-host", "localhost", "Database host")
	dbPort := flag.Int("db-port", 5432, "Database port")
	dbName := flag.String("db-name", "password_manager", "Database name")
	dbUser := flag.String("db-user", "postgres", "Database user")
	dbPassword := flag.String("db-password", "", "Database password")
	dbSSLMode := flag.String("db-ssl-mode", "disable", "Database SSL mode")

	flag.Parse()

	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		*dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSLMode)

	srv, err := server.NewServer(*host, *port, connStr)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
		os.Exit(1)
	}

	log.Printf("Starting server on %s:%d", *host, *port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}
