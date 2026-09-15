package handlers

import (
	"net/http"

	"ticket-system/internal/middleware"
	"ticket-system/internal/repository"
)

func NewRouter(store *repository.MemoryStore, jwtSecret []byte, authHandler *AuthHandler, ticketHandler *TicketHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	authMiddleware := middleware.Auth(jwtSecret)

	mux.Handle("POST /tickets", authMiddleware(http.HandlerFunc(ticketHandler.Create)))
	mux.Handle("GET /tickets", authMiddleware(http.HandlerFunc(ticketHandler.List)))
	mux.Handle("GET /tickets/{id}", authMiddleware(http.HandlerFunc(ticketHandler.Get)))
	mux.Handle("PATCH /tickets/{id}/status", authMiddleware(http.HandlerFunc(ticketHandler.UpdateStatus)))

	return mux
}
