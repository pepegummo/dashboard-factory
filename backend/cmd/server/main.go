package main

import (
	"log"
	"net/http"
	"os"

	"dashboard-builder/backend/internal/db"
	"dashboard-builder/backend/internal/httpapi"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/dashboard-builder.db"
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer conn.Close()

	router := httpapi.NewRouter(conn)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("dashboard-builder backend listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
