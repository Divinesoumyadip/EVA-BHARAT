package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"ticket-system/internal/handlers"
	"ticket-system/internal/repository"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	store := repository.NewMemoryStore()
	authHandler := handlers.NewAuthHandler(store, []byte(jwtSecret), 24*time.Hour)
	ticketHandler := handlers.NewTicketHandler(store)
	router := handlers.NewRouter(store, []byte(jwtSecret), authHandler, ticketHandler)

	addr := "0.0.0.0:" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
