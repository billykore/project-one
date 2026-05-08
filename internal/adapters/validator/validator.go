package validator

import "github.com/go-playground/validator/v10"

type Validator struct {
	v *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{v: validator.New()}
}

// Validate validates the given struct using go-playground/validator.
func (v *Validator) Validate(s any) error {
	return v.v.Struct(s)
}
