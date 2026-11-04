package domain

// Account is an account that can follow and be followed as social sees it. It is
// this context's own projection of an identity user and carries only the
// fields the context needs.
type Account struct {
	ID       int
	Username string
}
