# Design Spec: Notification Worker Refactoring

## 1. Objective
Refactor the notification background worker in `internal/adapters/notification/worker.go` to remove its dependency on the notification repository. Instead of persisting notifications directly, the background worker will consume notifications from the event broker and return them to the caller via a read-only channel (`<-chan *domain.Notification`). This separates event consumption from data persistence.

## 2. Proposed Changes

### 2.1. Core Ports Modification
Update the `NotificationConsumer` interface in `internal/core/ports/notification.go`:
```go
type NotificationConsumer interface {
	// Start begins the asynchronous consumption of events from the broker and returns a channel of notifications.
	Start(ctx context.Context) (<-chan *domain.Notification, error)
	// Stop gracefully terminates the background worker process.
	Stop(ctx context.Context) error
}
```

### 2.2. Background Worker Modification
Refactor `BackgroundWorker` in `internal/adapters/notification/worker.go`:
- Remove `repo ports.NotificationRepository` field from `BackgroundWorker` struct.
- Update `NewBackgroundWorker` constructor to:
  ```go
  func NewBackgroundWorker(ch <-chan *domain.Notification, log ports.Logger) *BackgroundWorker
  ```
- Update `Start` method:
  ```go
  func (w *BackgroundWorker) Start(ctx context.Context) (<-chan *domain.Notification, error) {
      w.mu.Lock()
      defer w.mu.Unlock()
      if w.started {
          return nil, ErrWorkerAlreadyStarted
      }
      w.started = true

      outCh := make(chan *domain.Notification, cap(w.ch))

      w.wg.Go(func() {
          w.log.Info(ctx, "background notification worker started")
          for notification := range w.ch {
              if notification == nil {
                  continue
              }
              outCh <- notification
          }
          close(outCh)
          w.log.Info(ctx, "notification channel closed, worker stopped cleanly")
      })
      return outCh, nil
  }
  ```

### 2.3. Application Entry Point Update
Update `cmd/main.go`:
- Instantiate the background worker without the repository:
  ```go
  worker := notificationBroker.NewBackgroundWorker(broker.Channel(), lgr)
  ```
- Receive the output channel from `worker.Start(ctx)`:
  ```go
  outCh, err := worker.Start(ctx)
  if err != nil {
      lgr.Fatal(ctx, "failed to start background notification worker", "error", err)
  }
  ```
- Start a separate background goroutine to process and persist notifications from `outCh`:
  ```go
  go func() {
      for n := range outCh {
          bgCtx := context.Background()
          if err := notificationRepo.Create(bgCtx, n); err != nil {
              lgr.Error(bgCtx, "failed to persist notification", "userID", n.UserID, "type", n.Type, "error", err)
          } else {
              lgr.Info(bgCtx, "notification persisted successfully", "id", n.ID, "userID", n.UserID)
          }
      }
  }()
  ```

### 2.4. Tests and Mocks Update
- Refactor `internal/adapters/notification/broker_worker_test.go` to conform to the updated constructor and consume the output channel asynchronously.
- Regenerate mocks by running `make mock`.

## 3. Verification Plan
- Run `go test ./...` to verify all tests (including mocks and adapters) pass.
- Run `go build ./...` to ensure compilation succeeds.
