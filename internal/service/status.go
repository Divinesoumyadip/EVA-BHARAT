package service

import (
	"errors"

	"ticket-system/internal/models"
)

var ErrInvalidTransition = errors.New("invalid status transition")

var allowedTransitions = map[models.TicketStatus]models.TicketStatus{
	models.StatusOpen:       models.StatusInProgress,
	models.StatusInProgress: models.StatusClosed,
}

func ValidateTransition(current, next models.TicketStatus) error {
	if !next.Valid() {
		return ErrInvalidTransition
	}
	allowed, exists := allowedTransitions[current]
	if !exists || allowed != next {
		return ErrInvalidTransition
	}
	return nil
}
