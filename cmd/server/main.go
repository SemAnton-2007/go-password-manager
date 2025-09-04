package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"password-manager/internal/server"
)

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
