package domain

import "errors"

// Sentinel errors owned by the featureflags context.
var (
	// ErrFlagNotFound is returned when a feature flag cannot be found.
	ErrFlagNotFound = errors.New("feature flag not found")
	// ErrFlagKeyExists is returned when creating a flag with a duplicate key.
	ErrFlagKeyExists = errors.New("feature flag key already exists")
	// ErrFlagArchived is returned when attempting to edit an archived flag.
	ErrFlagArchived = errors.New("feature flag is archived")
	// ErrRevisionConflict is returned when a flag update conflicts with a newer revision.
	ErrRevisionConflict = errors.New("feature flag revision conflict")
	// ErrInvalidFlagMode is returned when an availability mode is invalid.
	ErrInvalidFlagMode = errors.New("invalid feature flag mode")
	// ErrFlagSettingNotFound is returned when an environment setting does not exist.
	ErrFlagSettingNotFound = errors.New("feature flag setting not found")
	// ErrConflictingOverrides is returned when a user is both included and excluded.
	ErrConflictingOverrides = errors.New("user cannot be both included and excluded")
	// ErrOperatorOnly is returned when a non-operator attempts to administer flags.
	ErrOperatorOnly = errors.New("operator access required")
)
