package repository

import (
	"errors"
	"sync"
	"time"

	"ticket-system/internal/models"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUserExists     = errors.New("user already exists")
	ErrTicketNotFound = errors.New("ticket not found")
)

type MemoryStore struct {
	mu           sync.RWMutex
	users        map[int64]*models.User
	usersByEmail map[string]int64
	tickets      map[int64]*models.Ticket
	nextUserID   int64
	nextTicketID int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[int64]*models.User),
		usersByEmail: make(map[string]int64),
		tickets:      make(map[int64]*models.Ticket),
	}
}

func (s *MemoryStore) CreateUser(email, passwordHash string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[email]; exists {
		return nil, ErrUserExists
	}

	s.nextUserID++
	user := &models.User{
		ID:           s.nextUserID,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.users[user.ID] = user
	s.usersByEmail[email] = user.ID
	return user, nil
}

func (s *MemoryStore) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.usersByEmail[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return s.users[id], nil
}

func (s *MemoryStore) CreateTicket(userID int64, title, description string) (*models.Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextTicketID++
	now := time.Now()
	ticket := &models.Ticket{
		ID:          s.nextTicketID,
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      models.StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.tickets[ticket.ID] = ticket
	return ticket, nil
}

func (s *MemoryStore) ListTicketsByUser(userID int64) []*models.Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Ticket, 0)
	for _, ticket := range s.tickets {
		if ticket.UserID == userID {
			result = append(result, ticket)
		}
	}
	return result
}

func (s *MemoryStore) GetTicket(id int64) (*models.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, exists := s.tickets[id]
	if !exists {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

func (s *MemoryStore) UpdateTicketStatus(id int64, status models.TicketStatus) (*models.Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[id]
	if !exists {
		return nil, ErrTicketNotFound
	}
	ticket.Status = status
	ticket.UpdatedAt = time.Now()
	return ticket, nil
}
