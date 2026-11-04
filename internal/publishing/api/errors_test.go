package api

import (
	"testing"

	pkgmiddleware "github.com/billykore/project-one/internal/pkg/api/middleware"
	"github.com/stretchr/testify/assert"
)

func TestErrorMappingsResolveAfterRegistration(t *testing.T) {
	pkgmiddleware.RegisterErrorMappings(errorMappings)

	for sentinel, want := range errorMappings {
		assert.Equal(t, want, pkgmiddleware.LookupError(sentinel), sentinel.Error())
	}
}
