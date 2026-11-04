// Package auth contains transport-neutral authenticated-principal data.
package auth

import identitydomain "github.com/billykore/project-one/internal/identity/domain"

// Principal is the authenticated identity contract used outside identity.
// It aliases the existing model temporarily so callers can migrate without a
// flag day; downstream contexts must use only ID and Username.
type Principal = identitydomain.User
