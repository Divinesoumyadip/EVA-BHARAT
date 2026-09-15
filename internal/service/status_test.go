package service

import (
	"testing"

	"ticket-system/internal/models"
)

func TestValidateTransitionAllowed(t *testing.T) {
	cases := []struct {
		from models.TicketStatus
		to   models.TicketStatus
	}{
		{models.StatusOpen, models.StatusInProgress},
		{models.StatusInProgress, models.StatusClosed},
	}

	for _, c := range cases {
		if err := ValidateTransition(c.from, c.to); err != nil {
			t.Errorf("expected %s -> %s to be allowed, got error: %v", c.from, c.to, err)
		}
	}
}

func TestValidateTransitionRejected(t *testing.T) {
	cases := []struct {
		from models.TicketStatus
		to   models.TicketStatus
	}{
		{models.StatusOpen, models.StatusClosed},
		{models.StatusClosed, models.StatusOpen},
		{models.StatusClosed, models.StatusInProgress},
		{models.StatusInProgress, models.StatusOpen},
		{models.StatusOpen, models.StatusOpen},
		{models.StatusClosed, models.StatusClosed},
	}

	for _, c := range cases {
		if err := ValidateTransition(c.from, c.to); err != ErrInvalidTransition {
			t.Errorf("expected %s -> %s to be rejected, got: %v", c.from, c.to, err)
		}
	}
}

func TestValidateTransitionInvalidStatus(t *testing.T) {
	if err := ValidateTransition(models.StatusOpen, models.TicketStatus("bogus")); err != ErrInvalidTransition {
		t.Errorf("expected invalid status to be rejected, got: %v", err)
	}
}
