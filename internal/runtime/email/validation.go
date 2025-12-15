package email

import (
	"errors"

	"github.com/dimiro1/lunar/internal/services/email"
)

// ValidateSendRequest validates a SendRequest and returns an error if invalid.
func ValidateSendRequest(req email.SendRequest) error {
	if req.From == "" {
		return errors.New("from is required")
	}
	if len(req.To) == 0 {
		return errors.New("to is required")
	}
	if req.Subject == "" {
		return errors.New("subject is required")
	}
	if req.Text == "" && req.HTML == "" {
		return errors.New("either text or html content is required")
	}
	return nil
}
