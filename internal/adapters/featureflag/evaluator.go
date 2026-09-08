package featureflag

import (
	"context"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

// Evaluator evaluates feature flags from an in-memory environment snapshot.
type Evaluator struct {
	repository      ports.FeatureFlagRepository
	logger          ports.Logger
	environment     domain.Environment
	refreshInterval time.Duration

	mu       sync.RWMutex
	snapshot map[string]domain.FlagSnapshot
}

// NewEvaluator creates an evaluator and loads its first snapshot synchronously.
func NewEvaluator(
	repository ports.FeatureFlagRepository,
	logger ports.Logger,
	environment domain.Environment,
	refreshInterval time.Duration,
) (*Evaluator, error) {
	if repository == nil || logger == nil {
		panic("NewEvaluator: dependencies must not be nil")
	}
	if refreshInterval <= 0 {
		refreshInterval = 30 * time.Second
	}
	evaluator := &Evaluator{
		repository:      repository,
		logger:          logger,
		environment:     environment,
		refreshInterval: refreshInterval,
		snapshot:        make(map[string]domain.FlagSnapshot),
	}
	if err := evaluator.Refresh(context.Background()); err != nil {
		logger.Warn(context.Background(), "feature flag snapshot unavailable", "environment", environment, "failure_category", "snapshot_load")
	}
	return evaluator, nil
}

// StartRefreshLoop periodically reloads flag state until ctx is cancelled.
func (e *Evaluator) StartRefreshLoop(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(e.refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := e.Refresh(ctx); err != nil {
					e.logger.Warn(ctx, "feature flag snapshot refresh failed", "environment", e.environment, "failure_category", "snapshot_refresh", "error", err)
				}
			}
		}
	}()
}

// Refresh reloads all flags for the evaluator's environment.
func (e *Evaluator) Refresh(ctx context.Context) error {
	snapshots, err := e.repository.LoadSnapshot(ctx, e.environment)
	if err != nil {
		return err
	}
	next := make(map[string]domain.FlagSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		next[strings.ToLower(strings.TrimSpace(snapshot.Key))] = snapshot
	}
	e.mu.Lock()
	e.snapshot = next
	e.mu.Unlock()
	return nil
}

// Evaluate returns a decision without touching the database.
func (e *Evaluator) Evaluate(ctx context.Context, key, username string) domain.FeatureFlagDecision {
	normalizedKey := strings.ToLower(strings.TrimSpace(key))
	e.mu.RLock()
	snapshot, ok := e.snapshot[normalizedKey]
	e.mu.RUnlock()
	if !ok {
		return domain.FeatureFlagDecision{Key: key, Enabled: false, Source: domain.SourceUnknown, Reason: "flag not found"}
	}

	if snapshot.Lifecycle == domain.LifecycleArchived {
		return e.safeDefault(snapshot, domain.SourceArchived)
	}
	if snapshot.Setting == nil {
		return e.safeDefault(snapshot, domain.SourceSafeDefault)
	}
	if snapshot.Setting.Mode == domain.ModeDisabledAll {
		return domain.FeatureFlagDecision{Key: snapshot.Key, Enabled: false, Source: domain.SourceDisabledAll, Reason: "disabled for everyone"}
	}

	if username != "" {
		for _, override := range snapshot.Overrides {
			if override.Username != username {
				continue
			}
			if override.Type == domain.OverrideExclude {
				return domain.FeatureFlagDecision{Key: snapshot.Key, Enabled: false, Source: domain.SourceExclude, Reason: "explicitly excluded"}
			}
			if override.Type == domain.OverrideInclude {
				return domain.FeatureFlagDecision{Key: snapshot.Key, Enabled: true, Source: domain.SourceInclude, Reason: "explicitly included"}
			}
		}
	}

	switch snapshot.Setting.Mode {
	case domain.ModeGradual:
		if username == "" {
			return e.safeDefault(snapshot, domain.SourceSafeDefault)
		}
		bucket := rolloutBucket(snapshot.Key, e.environment, username)
		enabled := bucket < snapshot.Setting.RolloutPercentage
		return domain.FeatureFlagDecision{Key: snapshot.Key, Enabled: enabled, Source: domain.SourceGradual, Reason: "deterministic rollout allocation"}
	case domain.ModeEnabledAll:
		return domain.FeatureFlagDecision{Key: snapshot.Key, Enabled: true, Source: domain.SourceEnabledAll, Reason: "enabled for everyone"}
	default:
		e.logger.Warn(ctx, "feature flag evaluation failed", "flag", snapshot.Key, "environment", e.environment, "failure_category", "invalid_mode", "time", time.Now().UTC())
		return e.safeDefault(snapshot, domain.SourceSafeDefault)
	}
}

func (e *Evaluator) safeDefault(ctx domain.FlagSnapshot, source string) domain.FeatureFlagDecision {
	return domain.FeatureFlagDecision{Key: ctx.Key, Enabled: ctx.SafeDefault, Source: source, Reason: "safe default"}
}

func rolloutBucket(key string, environment domain.Environment, username string) int {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(key + "|" + string(environment) + "|" + username))
	return int(hash.Sum64() % 100)
}
