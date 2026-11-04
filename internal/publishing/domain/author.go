package domain

// Author is the account that owns posts and comments as publishing sees it. It is
// this context's own projection of an identity user and carries only the
// fields the context needs.
type Author struct {
	ID       int
	Username string
}
