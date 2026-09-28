package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"web-server-eccho/pkg/echo"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	handler := echo.NewHandler()

	log.Printf("Starting Echo Service on port %s", port)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start %v", err)
	}
}
