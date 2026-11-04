package ports

// Validator defines the interface for validating structs using go-playground/validator.
type Validator interface {
	// Validate validates the given struct and returns an error if validation fails.
	Validate(any) error
}
