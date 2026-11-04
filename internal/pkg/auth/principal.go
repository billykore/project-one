// Package auth contains transport-neutral authenticated-principal data.
package auth

// Principal is the authenticated identity contract used outside identity.
// The identity authorization middleware builds it from the session's user;
// downstream contexts see only the stable ID and the display username.
type Principal struct {
	ID       int
	Username string
}
