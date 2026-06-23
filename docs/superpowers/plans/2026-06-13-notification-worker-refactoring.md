# Notification Worker Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the notification background worker to act as a generic consumer returning a channel, moving the consumption and persistence logic to the usecase layer.

**Architecture:** The `BackgroundWorker` driving adapter consumes notifications from the event broker and exposes them via a read-only channel on `Start()`. The `NotificationUseCase` driving port is updated to orchestrate this consumption by starting the worker, reading from its channel, and writing to the database using `NotificationRepository`.

**Tech Stack:** Go 1.26+, GoMock, Testify, Clean Architecture.

---

### Task 1: Core Ports Modification

**Files:**
- Modify: `internal/core/ports/notification.go`

- [ ] **Step 1: Modify interfaces `NotificationConsumer` and `NotificationUseCase`**

Modify [internal/core/ports/notification.go](file:///Users/billykore/Kore/Golang/project1/internal/core/ports/notification.go#L23-L45) to match the new signatures:

```go
// NotificationUseCase is a driving port for notification business logic.
type NotificationUseCase interface {
	// GetNotifications retrieves notifications for the given user, resolved with actor details.
	GetNotifications(ctx context.Context, username string, limit, offset int) ([]*domain.NotificationDetail, error)
	// MarkAsRead verifies ownership and marks a specific notification as read.
	MarkAsRead(ctx context.Context, id int, username string) error
	// MarkAllAsRead marks all notifications as read for the authenticated user.
	MarkAllAsRead(ctx context.Context, username string) error

	// Start begins the background consumption and database persistence of notifications.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the background consumption.
	Stop(ctx context.Context) error
}

// NotificationConsumer is a driver port representing a background worker that consumes notification events.
type NotificationConsumer interface {
	// Start begins the asynchronous consumption of events from the broker.
	Start(ctx context.Context) (<-chan *domain.Notification, error)
	// Stop gracefully terminates the background worker process.
	Stop(ctx context.Context) error
}
```

- [ ] **Step 2: Verify compilation failure**

Run: `go build ./...`
Expected: Compilation failure because existing implementations do not satisfy the updated interfaces.

- [ ] **Step 3: Commit**

```bash
git add internal/core/ports/notification.go
git commit -m "refactor!: update NotificationConsumer and NotificationUseCase port signatures"
```

---

### Task 2: Background Worker Refactoring

**Files:**
- Modify: `internal/adapters/notification/worker.go`
- Modify: `internal/adapters/notification/broker_worker_test.go`

- [ ] **Step 1: Refactor `BackgroundWorker` and `NewBackgroundWorker`**

Modify [internal/adapters/notification/worker.go](file:///Users/billykore/Kore/Golang/project1/internal/adapters/notification/worker.go) to remove the repository dependency and return a channel on Start:

```go
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
	ch      <-chan *domain.Notification
	log     ports.Logger
	wg      sync.WaitGroup
	mu      sync.Mutex
	started bool
}

// NewBackgroundWorker creates a new BackgroundWorker instance.
func NewBackgroundWorker(ch <-chan *domain.Notification, log ports.Logger) *BackgroundWorker {
	return &BackgroundWorker{
		ch:  ch,
		log: log,
	}
}

// Start spawns the background worker goroutine to consume events.
// It returns ErrWorkerAlreadyStarted if the worker is already active.
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
```

- [ ] **Step 2: Update broker/worker unit test**

Modify [internal/adapters/notification/broker_worker_test.go](file:///Users/billykore/Kore/Golang/project1/internal/adapters/notification/broker_worker_test.go) to match the new constructor and Start signature:

```go
package notification

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	mu            sync.Mutex
	notifications []*domain.Notification
}

func (r *mockRepo) Create(ctx context.Context, n *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications = append(r.notifications, n)
	return nil
}

func (r *mockRepo) GetByID(ctx context.Context, id int) (*domain.Notification, error) {
	return nil, nil
}

func (r *mockRepo) GetByUserID(ctx context.Context, u int, l, o int) ([]*domain.Notification, error) {
	return nil, nil
}

func (r *mockRepo) MarkAsRead(ctx context.Context, id int) error   { return nil }
func (r *mockRepo) MarkAllAsRead(ctx context.Context, u int) error { return nil }

type mockLogger struct{}

func (mockLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (mockLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (mockLogger) Warn(ctx context.Context, msg string, fields ...any)  {}
func (mockLogger) Error(ctx context.Context, msg string, fields ...any) {}
func (mockLogger) Fatal(ctx context.Context, msg string, fields ...any) {}

func TestBrokerAndWorker(t *testing.T) {
	broker := NewMemoryBroker(10)
	repo := &mockRepo{}
	worker := NewBackgroundWorker(broker.Channel(), mockLogger{})

	ctx := context.Background()
	outCh, err := worker.Start(ctx)
	assert.NoError(t, err)

	n := &domain.Notification{UserID: 1, ActorID: 2, Type: domain.NotificationTypeFollow}
	err = broker.Publish(ctx, n)
	assert.NoError(t, err)

	// Consume and save notifications in a mock consumer routine
	go func() {
		for notification := range outCh {
			_ = repo.Create(ctx, notification)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	repo.mu.Lock()
	count := len(repo.notifications)
	repo.mu.Unlock()
	assert.Equal(t, 1, count)

	broker.Close()
	err = worker.Stop(ctx)
	assert.NoError(t, err)
}
```

- [ ] **Step 3: Run notification adapter tests**

Run: `go test -v ./internal/adapters/notification/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/adapters/notification/worker.go internal/adapters/notification/broker_worker_test.go
git commit -m "feat: refactor BackgroundWorker to return channel and remove repo dependency"
```

---

### Task 3: Notification UseCase Modification

**Files:**
- Modify: `internal/core/usecase/notification_usecase.go`
- Modify: `internal/core/usecase/notification_usecase_test.go`

- [ ] **Step 1: Refactor `notificationUseCase` struct and `NewNotificationUseCase`**

Modify [internal/core/usecase/notification_usecase.go](file:///Users/billykore/Kore/Golang/project1/internal/core/usecase/notification_usecase.go):

```go
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
)

type notificationUseCase struct {
	repo     ports.NotificationRepository
	userRepo ports.UserRepository
	consumer ports.NotificationConsumer
	log      ports.Logger
}

func NewNotificationUseCase(
	repo ports.NotificationRepository,
	userRepo ports.UserRepository,
	consumer ports.NotificationConsumer,
	log ports.Logger,
) ports.NotificationUseCase {
	if repo == nil {
		panic("NewNotificationUseCase: repo is required")
	}
	if userRepo == nil {
		panic("NewNotificationUseCase: userRepo is required")
	}
	if consumer == nil {
		panic("NewNotificationUseCase: consumer is required")
	}
	if log == nil {
		panic("NewNotificationUseCase: log is required")
	}
	return &notificationUseCase{
		repo:     repo,
		userRepo: userRepo,
		consumer: consumer,
		log:      log,
	}
}

func (uc *notificationUseCase) Start(ctx context.Context) error {
	outCh, err := uc.consumer.Start(ctx)
	if err != nil {
		return err
	}

	go func() {
		for n := range outCh {
			if n == nil {
				continue
			}
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

func (uc *notificationUseCase) Stop(ctx context.Context) error {
	return uc.consumer.Stop(ctx)
}

func (uc *notificationUseCase) GetNotifications(ctx context.Context, username string, limit, offset int) ([]*domain.NotificationDetail, error) {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("get user by username: %w", domain.ErrUserNotFound)
	}
	notifications, err := uc.repo.GetByUserID(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get notifications by user id: %w", err)
	}

	actorMap := map[int]string{
		user.ID: user.Username,
	}
	details := make([]*domain.NotificationDetail, 0, len(notifications))
	for _, n := range notifications {
		if n == nil {
			continue
		}
		actorUsername, exists := actorMap[n.ActorID]
		if !exists {
			actor, err := uc.userRepo.GetUserByID(ctx, n.ActorID)
			if err != nil {
				if errors.Is(err, domain.ErrUserNotFound) {
					actorUsername = ""
					actorMap[n.ActorID] = ""
				} else {
					return nil, fmt.Errorf("get actor by id: %w", err)
				}
			} else {
				if actor != nil {
					actorUsername = actor.Username
				} else {
					actorUsername = ""
				}
				actorMap[n.ActorID] = actorUsername
			}
		}
		details = append(details, &domain.NotificationDetail{
			Notification:  *n,
			ActorUsername: actorUsername,
		})
	}
	return details, nil
}

func (uc *notificationUseCase) MarkAsRead(ctx context.Context, id int, username string) error {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return fmt.Errorf("get user by username: %w", domain.ErrUserNotFound)
	}
	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get notification by id: %w", err)
	}
	if notification == nil {
		return domain.ErrNotificationNotFound
	}
	if notification.UserID != user.ID {
		return domain.ErrUnauthorized
	}
	err = uc.repo.MarkAsRead(ctx, id)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}
	return nil
}

func (uc *notificationUseCase) MarkAllAsRead(ctx context.Context, username string) error {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("get user by username: %w", err)
	}
	if user == nil {
		return fmt.Errorf("get user by username: %w", domain.ErrUserNotFound)
	}
	err = uc.repo.MarkAllAsRead(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Update UseCase unit tests**

Modify [internal/core/usecase/notification_usecase_test.go](file:///Users/billykore/Kore/Golang/project1/internal/core/usecase/notification_usecase_test.go):

```go
package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type mockUseCaseLogger struct{}

func (mockUseCaseLogger) Debug(ctx context.Context, msg string, fields ...any) {}
func (mockUseCaseLogger) Info(ctx context.Context, msg string, fields ...any)  {}
func (mockUseCaseLogger) Warn(ctx context.Context, msg string, fields ...any)  {}
func (mockUseCaseLogger) Error(ctx context.Context, msg string, fields ...any) {}
func (mockUseCaseLogger) Fatal(ctx context.Context, msg string, fields ...any) {}

func TestNewNotificationUseCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	lgr := mockUseCaseLogger{}

	t.Run("success", func(t *testing.T) {
		uc := NewNotificationUseCase(mockRepo, mockUserRepo, mockConsumer, lgr)
		assert.NotNil(t, uc)
	})

	t.Run("nil repo", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: repo is required", func() {
			NewNotificationUseCase(nil, mockUserRepo, mockConsumer, lgr)
		})
	})

	t.Run("nil userRepo", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: userRepo is required", func() {
			NewNotificationUseCase(mockRepo, nil, mockConsumer, lgr)
		})
	})

	t.Run("nil consumer", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: consumer is required", func() {
			NewNotificationUseCase(mockRepo, mockUserRepo, nil, lgr)
		})
	})

	t.Run("nil logger", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: log is required", func() {
			NewNotificationUseCase(mockRepo, mockUserRepo, mockConsumer, nil)
		})
	})
}

func TestNotificationUseCase_Lifecycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	lgr := mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mockConsumer, lgr)

	ctx := context.Background()

	t.Run("Start success and consume", func(t *testing.T) {
		ch := make(chan *domain.Notification, 1)
		mockConsumer.EXPECT().Start(ctx).Return(ch, nil)

		notification := &domain.Notification{ID: 101, UserID: 1}
		mockRepo.EXPECT().Create(gomock.Any(), notification).Return(nil)

		err := uc.Start(ctx)
		assert.NoError(t, err)

		ch <- notification
		time.Sleep(50 * time.Millisecond) // yield to allow the goroutine to run
	})

	t.Run("Stop success", func(t *testing.T) {
		mockConsumer.EXPECT().Stop(ctx).Return(nil)
		err := uc.Stop(ctx)
		assert.NoError(t, err)
	})
}

func TestNotificationUseCase_GetNotifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	lgr := mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mockConsumer, lgr)

	ctx := context.Background()
	username := "testuser"
	limit := 10
	offset := 0

	user := &domain.User{
		ID:       1,
		Username: username,
	}

	t.Run("success with caching and ignored not found actor lookups", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)

		notifications := []*domain.Notification{
			{ID: 101, UserID: 1, ActorID: 2, Type: domain.NotificationTypeFollow},
			{ID: 102, UserID: 1, ActorID: 2, Type: domain.NotificationTypeLike},
			{ID: 103, UserID: 1, ActorID: 3, Type: domain.NotificationTypeComment},
			{ID: 104, UserID: 1, ActorID: 4, Type: domain.NotificationTypeComment},
			{ID: 105, UserID: 1, ActorID: 4, Type: domain.NotificationTypeComment}, // Second notification from the same missing actor ID 4
			nil, // Nil notification to test skipping without dereferencing/panicking
			{ID: 106, UserID: 1, ActorID: 1, Type: domain.NotificationTypeComment}, // Notification where actor is the current user (ID 1), should use pre-populated cache
		}
		mockRepo.EXPECT().GetByUserID(ctx, user.ID, limit, offset).Return(notifications, nil)

		// Actor 2: lookup succeeds once
		actor2 := &domain.User{ID: 2, Username: "actor2"}
		mockUserRepo.EXPECT().GetUserByID(ctx, 2).Return(actor2, nil).Times(1)

		// Actor 3: lookup succeeds
		actor3 := &domain.User{ID: 3, Username: "actor3"}
		mockUserRepo.EXPECT().GetUserByID(ctx, 3).Return(actor3, nil)

		// Actor 4: lookup fails with ErrUserNotFound, should only be called once due to caching of soft failure
		mockUserRepo.EXPECT().GetUserByID(ctx, 4).Return(nil, domain.ErrUserNotFound).Times(1)

		results, err := uc.GetNotifications(ctx, username, limit, offset)
		assert.NoError(t, err)
		assert.Len(t, results, 6)

		assert.Equal(t, "actor2", results[0].ActorUsername)
		assert.Equal(t, "actor2", results[1].ActorUsername)
		assert.Equal(t, "actor3", results[2].ActorUsername)
		assert.Equal(t, "", results[3].ActorUsername)
		assert.Equal(t, "", results[4].ActorUsername)
		assert.Equal(t, username, results[5].ActorUsername)
	})

	t.Run("user repo error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, expectedErr)

		results, err := uc.GetNotifications(ctx, username, limit, offset)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, results)
	})

	t.Run("nil user from repo", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, nil)

		results, err := uc.GetNotifications(ctx, username, limit, offset)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, results)
	})

	t.Run("notification repo error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		mockRepo.EXPECT().GetByUserID(ctx, user.ID, limit, offset).Return(nil, expectedErr)

		results, err := uc.GetNotifications(ctx, username, limit, offset)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, results)
	})

	t.Run("actor lookup generic error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)

		notifications := []*domain.Notification{
			{ID: 101, UserID: 1, ActorID: 5, Type: domain.NotificationTypeFollow},
		}
		mockRepo.EXPECT().GetByUserID(ctx, user.ID, limit, offset).Return(notifications, nil)

		expectedErr := errors.New("connection failed")
		mockUserRepo.EXPECT().GetUserByID(ctx, 5).Return(nil, expectedErr)

		results, err := uc.GetNotifications(ctx, username, limit, offset)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, results)
	})
}

func TestNotificationUseCase_MarkAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	lgr := mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mockConsumer, lgr)

	ctx := context.Background()
	username := "testuser"
	notificationID := 101

	user := &domain.User{
		ID:       1,
		Username: username,
	}

	t.Run("success", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		notification := &domain.Notification{ID: notificationID, UserID: user.ID}
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(notification, nil)
		mockRepo.EXPECT().MarkAsRead(ctx, notificationID).Return(nil)

		err := uc.MarkAsRead(ctx, notificationID, username)
		assert.NoError(t, err)
	})

	t.Run("user repo error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, expectedErr)

		err := uc.MarkAsRead(ctx, notificationID, username)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("nil user from repo", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, nil)

		err := uc.MarkAsRead(ctx, notificationID, username)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("notification repo get error", func(t *testing.T) {
		expectedErr := errors.New("not found")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(nil, expectedErr)

		err := uc.MarkAsRead(ctx, notificationID, username)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("notification nil check", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(nil, nil)

		err := uc.MarkAsRead(ctx, notificationID, username)
		assert.ErrorIs(t, err, domain.ErrNotificationNotFound)
	})

	t.Run("unauthorized owner mismatch", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		notification := &domain.Notification{ID: notificationID, UserID: 999}
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(notification, nil)

		err := uc.MarkAsRead(ctx, notificationID, username)
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("mark as read repository error", func(t *testing.T) {
		expectedErr := errors.New("db write error")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		notification := &domain.Notification{ID: notificationID, UserID: user.ID}
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(notification, nil)
		mockRepo.EXPECT().MarkAsRead(ctx, notificationID).Return(expectedErr)

		err := uc.MarkAsRead(ctx, notificationID, username)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestNotificationUseCase_MarkAllAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	lgr := mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mockConsumer, lgr)

	ctx := context.Background()
	username := "testuser"

	user := &domain.User{
		ID:       1,
		Username: username,
	}

	t.Run("success", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		mockRepo.EXPECT().MarkAllAsRead(ctx, user.ID).Return(nil)

		err := uc.MarkAllAsRead(ctx, username)
		assert.NoError(t, err)
	})

	t.Run("user repo error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, expectedErr)

		err := uc.MarkAllAsRead(ctx, username)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("nil user from repo", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, nil)

		err := uc.MarkAllAsRead(ctx, username)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("mark all as read repository error", func(t *testing.T) {
		expectedErr := errors.New("db write error")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
		mockRepo.EXPECT().MarkAllAsRead(ctx, user.ID).Return(expectedErr)

		err := uc.MarkAllAsRead(ctx, username)
		assert.ErrorIs(t, err, expectedErr)
	})
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/core/usecase/notification_usecase.go internal/core/usecase/notification_usecase_test.go
git commit -m "feat: implement Start and Stop in NotificationUseCase and update tests"
```

---

### Task 4: Mocks Regeneration

- [ ] **Step 1: Regenerate mocks**

Run: `make mock`
Expected: Successfully generates updated mock file in `internal/core/usecase/mocks/mock_notification.go` satisfying the new interfaces.

- [ ] **Step 2: Commit**

```bash
git add internal/core/usecase/mocks/mock_notification.go
git commit -m "chore: regenerate notification usecase and consumer mocks"
```

---

### Task 5: Main Entry Point Update

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: Update background worker & notification usecase instantiation and startup**

Modify [cmd/main.go](file:///Users/billykore/Kore/Golang/project1/cmd/main.go):

Update initialization logic (around line 74-87):
```go
	broker := notificationBroker.NewMemoryBroker(100)
	notificationRepo := repository.NewNotificationRepository(db)
	worker := notificationBroker.NewBackgroundWorker(broker.Channel(), lgr)
```

Update UseCase initialization & Start logic:
```go
	// 4. Initialize UseCase.
	loginUc := usecase.NewLoginUseCase(userRepo, tokenSvc, userTokenRepo, hasher, lgr)
	userUc := usecase.NewUserUseCase(userRepo, userTokenRepo, hasher)
	postUc := usecase.NewPostUseCase(postRepo, likeRepo, userRepo, broker, lgr)
	followUc := usecase.NewFollowUseCase(followRepo, userRepo, broker, lgr)
	commentUc := usecase.NewCommentUseCase(commentRepo, postRepo, userRepo, broker, lgr)
	notificationUc := usecase.NewNotificationUseCase(notificationRepo, userRepo, worker, lgr)

	if err := notificationUc.Start(ctx); err != nil {
		lgr.Fatal(ctx, "failed to start background notification usecase", "error", err)
	}
```

Update shutdown logic (around line 184):
```go
	broker.Close()
	if err := notificationUc.Stop(ctxShutdown); err != nil {
		lgr.Error(ctxShutdown, "failed to stop background notification usecase gracefully", "error", err)
	}
```

- [ ] **Step 2: Run all checks & build**

Run: `make check`
Expected: docs, vet, lint, and test pass.

Run: `make build`
Expected: Compile successfully.

- [ ] **Step 3: Commit**

```bash
git add cmd/main.go
git commit -m "feat: initialize and manage NotificationUseCase in main entrypoint"
```
