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
	repo     ports.NotificationRepository
	ch       <-chan *domain.Notification
	log      ports.Logger
	wg       sync.WaitGroup
	quit     chan struct{}
	stopOnce sync.Once
	mu       sync.Mutex
	started  bool
}

// NewBackgroundWorker creates a new BackgroundWorker instance.
func NewBackgroundWorker(repo ports.NotificationRepository, ch <-chan *domain.Notification, log ports.Logger) *BackgroundWorker {
	return &BackgroundWorker{
		repo: repo,
		ch:   ch,
		log:  log,
		quit: make(chan struct{}),
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
		for {
			select {
			case notification, ok := <-w.ch:
				if !ok {
					w.log.Info(ctx, "notification channel closed, worker stopping")
					return
				}
				if err := w.repo.Create(ctx, notification); err != nil {
					w.log.Error(ctx, "failed to persist notification", "userID", notification.UserID, "type", notification.Type, "error", err)
				} else {
					w.log.Info(ctx, "notification persisted successfully", "id", notification.ID, "userID", notification.UserID)
				}
			case <-w.quit:
				w.log.Info(ctx, "stop signal received, background worker exiting")
				// Drain any remaining notifications in the channel buffer non-blockingly
				for {
					select {
					case notification, ok := <-w.ch:
						if !ok {
							return
						}
						if err := w.repo.Create(ctx, notification); err != nil {
							w.log.Error(ctx, "failed to persist notification on drain", "userID", notification.UserID, "type", notification.Type, "error", err)
						} else {
							w.log.Info(ctx, "notification persisted successfully on drain", "id", notification.ID, "userID", notification.UserID)
						}
					default:
						return
					}
				}
			}
		}
	})
	return nil
}

// Stop gracefully shuts down the background worker.
// It waits for the worker loop to exit or returns the context error if ctx is cancelled first.
func (w *BackgroundWorker) Stop(ctx context.Context) error {
	w.stopOnce.Do(func() {
		close(w.quit)
	})

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
