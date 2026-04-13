package domain

import "errors"

// Greeting is the core domain entity representing a greeting message.
type Greeting struct {
	Message string
}

// Validate performs domain-level validation on the Greeting entity.
func (g *Greeting) Validate() error {
	if g.Message == "" {
		return errors.New("message is required")
	}
	return nil
}
