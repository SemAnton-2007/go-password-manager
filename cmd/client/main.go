// Package main предоставляет клиентское приложение для менеджера паролей.
package main

import (
	"flag"
	"fmt"
	"os"

	"password-manager/internal/client"
)

var (
	version   = "1.0.1"
	buildDate = "2025-09-06"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Password Manager Client\nVersion: %s\nBuild Date: %s\n", version, buildDate)
		return
	}

	host := flag.String("host", "", "Server host (optional)")
	port := flag.Int("port", 0, "Server port (optional)")
	flag.Parse()

	uiClient := client.NewUIClient(*host, *port)
	if err := uiClient.Run(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}
}
