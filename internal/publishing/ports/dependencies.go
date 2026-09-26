package ports

import (
	"context"

	featureflagdomain "github.com/billykore/project-one/internal/featureflags/domain"
	identitydomain "github.com/billykore/project-one/internal/identity/domain"
)

// UserLookup is publishing's narrow view of the identity context.
type UserLookup interface {
	GetUserByID(ctx context.Context, id int) (*identitydomain.User, error)
}

// FeatureEvaluator is publishing's release-control dependency. It exposes only
// the feature-flags published decision language, not administration or storage.
type FeatureEvaluator interface {
	Evaluate(ctx context.Context, key, username string) featureflagdomain.FeatureFlagDecision
}
