package ports

import (
	"context"

	"github.com/billykore/project-one/internal/operations/domain"
)

// DeliveryBacklogReader is a driven port that reads the event delivery backlog.
type DeliveryBacklogReader interface {
	// DeliveryBacklog returns the current backlog. Implementations must honour
	// the context deadline so a metrics scrape stays bounded.
	DeliveryBacklog(ctx context.Context) (domain.DeliveryBacklog, error)
}
