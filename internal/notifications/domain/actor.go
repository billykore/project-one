package domain

// Actor is the account that triggered a notification as notifications sees it. It is
// this context's own projection of an identity user and carries only the
// fields the context needs.
type Actor struct {
	ID       int
	Username string
}
