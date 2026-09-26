// Package health adapts external dependencies to the core dependency-checker port.
package health

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/operations/ports"
	platformports "github.com/billykore/project-one/internal/platform/ports"
)

// componentDatabase is the stable component name of the application store.
const componentDatabase = "database"

// Pinger is the minimal database capability a readiness check needs.
type Pinger interface {
	// PingContext verifies the database connection is usable.
	PingContext(ctx context.Context) error
}

type databaseChecker struct {
	pinger Pinger
}

// NewDatabaseChecker creates a dependency checker backed by a database ping.
// The checked component is always reported as "database".
func NewDatabaseChecker(pinger Pinger) ports.DependencyChecker {
	return &databaseChecker{pinger: pinger}
}

func (c *databaseChecker) Name() string { return componentDatabase }

func (c *databaseChecker) Check(ctx context.Context) error {
	if c.pinger == nil {
		return errors.New("database health check is not configured")
	}
	if err := c.pinger.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping: %w", err)
	}
	return nil
}

type notificationChecker struct {
	name     string
	reporter platformports.HealthReporter
}

// NewNotificationChecker creates a dependency checker for notification delivery.
// It reads the broker client's connection state instead of publishing a
// synthetic event, so readiness never mutates delivery state. The subscriber is
// used rather than the publisher because the publisher connects lazily.
func NewNotificationChecker(name string, reporter platformports.HealthReporter) ports.DependencyChecker {
	if name == "" {
		name = "broker"
	}
	return &notificationChecker{name: name, reporter: reporter}
}

func (c *notificationChecker) Name() string { return c.name }

func (c *notificationChecker) Check(_ context.Context) error {
	if c.reporter == nil {
		return errors.New("notification subscriber health is not configured")
	}
	if !c.reporter.Healthy() {
		return errors.New("notification subscriber is not connected")
	}
	return nil
}
