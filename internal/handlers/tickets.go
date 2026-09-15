package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ticket-system/internal/middleware"
	"ticket-system/internal/models"
	"ticket-system/internal/repository"
	"ticket-system/internal/service"
)

type TicketHandler struct {
	store *repository.MemoryStore
}

func NewTicketHandler(store *repository.MemoryStore) *TicketHandler {
	return &TicketHandler{store: store}
}

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	ticket, err := h.store.CreateTicket(userID, title, strings.TrimSpace(req.Description))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create ticket")
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tickets := h.store.ListTicketsByUser(userID)
	writeJSON(w, http.StatusOK, tickets)
}

func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	ticket, err := h.store.GetTicket(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	if ticket.UserID != userID {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newStatus := models.TicketStatus(req.Status)
	if !newStatus.Valid() {
		writeError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	ticket, err := h.store.GetTicket(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	if ticket.UserID != userID {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	if err := service.ValidateTransition(ticket.Status, newStatus); err != nil {
		writeError(w, http.StatusBadRequest, "invalid status transition")
		return
	}

	updated, err := h.store.UpdateTicketStatus(id, newStatus)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update ticket")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}
