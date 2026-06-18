# Design Spec: Notification Worker Refactoring

## 1. Objective
Refactor the notification background worker in `internal/adapters/notification/worker.go` to remove its dependency on the notification repository. Instead of persisting notifications directly, the background worker will consume notifications from the event broker and return them to the caller via a read-only channel (`<-chan *domain.Notification`).
The logic to consume and persist notifications will be moved to the usecase layer in `internal/core/usecase/notification_usecase.go`, ensuring that the worker remains a generic consumer adapter, and the orchestration/persistence logic remains in the core usecase.

## 2. Proposed Changes

### 2.1. Core Ports Modification
Update the interfaces in `internal/core/ports/notification.go`:

1. `NotificationConsumer` - returns a read-only channel on Start:
```go
type NotificationConsumer interface {
	// Start begins the asynchronous consumption of events from the broker and returns a channel of notifications.
	Start(ctx context.Context) (<-chan *domain.Notification, error)
	// Stop gracefully terminates the background worker process.
	Stop(ctx context.Context) error
}
```

2. `NotificationUseCase` - adds Start and Stop methods for lifecycle management:
```go
type NotificationUseCase interface {
	GetNotifications(ctx context.Context, username string, limit, offset int) ([]*domain.NotificationDetail, error)
	MarkAsRead(ctx context.Context, id int, username string) error
	MarkAllAsRead(ctx context.Context, username string) error

	// Start begins the background consumption and database persistence of notifications.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the background consumption.
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

### 2.3. Notification UseCase Modification
Refactor `notificationUseCase` in `internal/core/usecase/notification_usecase.go`:
- Add fields for `consumer ports.NotificationConsumer` and `log ports.Logger`.
- Update `NewNotificationUseCase` constructor to:
  ```go
  func NewNotificationUseCase(
      repo ports.NotificationRepository,
      userRepo ports.UserRepository,
      consumer ports.NotificationConsumer,
      log ports.Logger,
  ) ports.NotificationUseCase
  ```
- Implement `Start(ctx context.Context) error` method:
  ```go
  func (uc *notificationUseCase) Start(ctx context.Context) error {
      outCh, err := uc.consumer.Start(ctx)
      if err != nil {
          return err
      }

      go func() {
          for n := range outCh {
              bgCtx := context.Background()
              if err := uc.repo.Create(bgCtx, n); err != nil {
                  uc.log.Error(bgCtx, "failed to persist notification", "userID", n.UserID, "type", n.Type, "error", err)
              } else {
                  uc.log.Info(bgCtx, "notification persisted successfully", "id", n.ID, "userID", n.UserID)
              }
          }
      }()

      return nil
  }
  ```
- Implement `Stop(ctx context.Context) error` method:
  ```go
  func (uc *notificationUseCase) Stop(ctx context.Context) error {
      return uc.consumer.Stop(ctx)
  }
  ```

### 2.4. Application Entry Point Update
Update `cmd/main.go`:
- Instantiate the background worker without the repository:
  ```go
  worker := notificationBroker.NewBackgroundWorker(broker.Channel(), lgr)
  ```
- Instantiate `NotificationUseCase` including the worker and logger:
  ```go
  notificationUc := usecase.NewNotificationUseCase(notificationRepo, userRepo, worker, lgr)
  ```
- Start the usecase:
  ```go
  if err := notificationUc.Start(ctx); err != nil {
      lgr.Fatal(ctx, "failed to start background notification usecase", "error", err)
  }
  ```
- On shutdown, stop the usecase:
  ```go
  if err := notificationUc.Stop(ctxShutdown); err != nil {
      lgr.Error(ctxShutdown, "failed to stop notification usecase gracefully", "error", err)
  }
  ```

### 2.5. Tests and Mocks Update
- Refactor `internal/adapters/notification/broker_worker_test.go` to conform to the updated constructor and consume the output channel asynchronously.
- Update `internal/core/usecase/notification_usecase_test.go` to test the updated `NewNotificationUseCase` constructor, `Start`, and `Stop` behaviors.
- Regenerate mocks by running `make mock`.

## 3. Verification Plan
- Run `go test ./...` to verify all tests (including mocks, usecase, and adapters) pass.
- Run `go build ./...` to ensure compilation succeeds.
