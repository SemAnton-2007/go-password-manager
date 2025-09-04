package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

type Server struct {
	host     string
	port     int
	database *Database
}

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

func (s *Server) Stop() error {
	return s.database.Close()
}
