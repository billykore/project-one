package notification

import (
	"context"
	"sync"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

var _ ports.NotificationConsumer = (*BackgroundWorker)(nil)

type BackgroundWorker struct {
	repo ports.NotificationRepository
	ch   <-chan *domain.Notification
	log  ports.Logger
	wg   sync.WaitGroup
	quit chan struct{}
}

func NewBackgroundWorker(repo ports.NotificationRepository, ch <-chan *domain.Notification, log ports.Logger) *BackgroundWorker {
	return &BackgroundWorker{
		repo: repo,
		ch:   ch,
		log:  log,
		quit: make(chan struct{}),
	}
}

func (w *BackgroundWorker) Start(ctx context.Context) error {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.log.Info(context.Background(), "background notification worker started")
		for {
			select {
			case notification, ok := <-w.ch:
				if !ok {
					w.log.Info(context.Background(), "notification channel closed, worker stopping")
					return
				}
				bgCtx := context.Background()
				if err := w.repo.Create(bgCtx, notification); err != nil {
					w.log.Error(bgCtx, "failed to persist notification", "userID", notification.UserID, "type", notification.Type, "error", err)
				} else {
					w.log.Info(bgCtx, "notification persisted successfully", "id", notification.ID, "userID", notification.UserID)
				}
			case <-w.quit:
				w.log.Info(context.Background(), "stop signal received, background worker exiting")
				return
			}
		}
	}()
	return nil
}

func (w *BackgroundWorker) Stop(ctx context.Context) error {
	close(w.quit)
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
