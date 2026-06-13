package notification

import (
	"context"
	"errors"
	"sync"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

var _ ports.NotificationConsumer = (*BackgroundWorker)(nil)

// ErrWorkerAlreadyStarted is returned when attempting to start a worker that is already running.
var ErrWorkerAlreadyStarted = errors.New("background worker already started")

// BackgroundWorker is a consumer of notification events that processes and persists
// incoming notifications in the database in the background.
type BackgroundWorker struct {
	repo    ports.NotificationRepository
	ch      <-chan *domain.Notification
	log     ports.Logger
	wg      sync.WaitGroup
	mu      sync.Mutex
	started bool
}

// NewBackgroundWorker creates a new BackgroundWorker instance.
func NewBackgroundWorker(repo ports.NotificationRepository, ch <-chan *domain.Notification, log ports.Logger) *BackgroundWorker {
	return &BackgroundWorker{
		repo: repo,
		ch:   ch,
		log:  log,
	}
}

// Start spawns the background worker goroutine to consume events.
// It returns ErrWorkerAlreadyStarted if the worker is already active.
func (w *BackgroundWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.started {
		return ErrWorkerAlreadyStarted
	}
	w.started = true

	w.wg.Go(func() {
		w.log.Info(ctx, "background notification worker started")
		for notification := range w.ch {
			if notification == nil {
				continue
			}
			bgCtx := context.Background()
			if err := w.repo.Create(bgCtx, notification); err != nil {
				w.log.Error(bgCtx, "failed to persist notification", "userID", notification.UserID, "type", notification.Type, "error", err)
			} else {
				w.log.Info(bgCtx, "notification persisted successfully", "id", notification.ID, "userID", notification.UserID)
			}
		}
		w.log.Info(ctx, "notification channel closed, worker stopped cleanly")
	})
	return nil
}

// Stop gracefully shuts down the background worker.
// It waits for the worker loop to exit or returns the context error if ctx is cancelled first.
func (w *BackgroundWorker) Stop(ctx context.Context) error {
	c := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		w.log.Info(ctx, "background worker stopped cleanly")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
