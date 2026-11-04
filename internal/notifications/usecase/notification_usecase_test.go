package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/notifications/domain"
	"github.com/billykore/project-one/internal/notifications/ports"
	"github.com/billykore/project-one/internal/pkg/auth"
	"github.com/billykore/project-one/internal/pkg/problem"
	"github.com/billykore/project-one/internal/testkit/mocks"
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
	mockUserRepo := mocks.NewMockActorLookup(ctrl)
	lgr := &mockUseCaseLogger{}

	t.Run("success", func(t *testing.T) {
		uc := NewNotificationUseCase(mockRepo, mockUserRepo, mocks.NewMockRealtimeNotifier(ctrl), lgr)
		assert.NotNil(t, uc)
	})

	t.Run("nil repo", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: dependencies must not be nil", func() {
			NewNotificationUseCase(nil, mockUserRepo, mocks.NewMockRealtimeNotifier(ctrl), lgr)
		})
	})

	t.Run("nil userRepo", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: dependencies must not be nil", func() {
			NewNotificationUseCase(mockRepo, nil, mocks.NewMockRealtimeNotifier(ctrl), lgr)
		})
	})

	t.Run("nil notifier", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: dependencies must not be nil", func() {
			NewNotificationUseCase(mockRepo, mockUserRepo, nil, lgr)
		})
	})

	t.Run("nil logger", func(t *testing.T) {
		assert.PanicsWithValue(t, "NewNotificationUseCase: dependencies must not be nil", func() {
			NewNotificationUseCase(mockRepo, mockUserRepo, mocks.NewMockRealtimeNotifier(ctrl), nil)
		})
	})
}

func TestNotificationUseCase_RecordEvent(t *testing.T) {
	ctx := context.Background()
	event := domain.NotificationEvent{
		EventID:       "backend-1",
		SchemaVersion: domain.NotificationEventSchemaVersion,
		Notification:  domain.Notification{UserID: 1, ActorID: 2, Type: domain.NotificationTypeFollow},
	}
	stored := func(n *domain.Notification) bool { return n.EventID == "backend-1" && n.UserID == 1 }

	setup := func(t *testing.T) (*mocks.MockNotificationRepository, *mocks.MockRealtimeNotifier, ports.NotificationUseCase) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockNotificationRepository(ctrl)
		notifier := mocks.NewMockRealtimeNotifier(ctrl)
		return repo, notifier, NewNotificationUseCase(repo, mocks.NewMockActorLookup(ctrl), notifier, &mockUseCaseLogger{})
	}

	t.Run("stores the event and pushes it", func(t *testing.T) {
		repo, notifier, uc := setup(t)
		repo.EXPECT().Create(ctx, gomock.Cond(stored)).Return(nil)
		notifier.EXPECT().Notify(ctx, gomock.Any()).Return(true, nil)
		assert.NoError(t, uc.RecordEvent(ctx, event))
	})

	t.Run("redelivered event is acknowledged without a push", func(t *testing.T) {
		repo, _, uc := setup(t)
		repo.EXPECT().Create(ctx, gomock.Any()).Return(domain.ErrDuplicateNotification)
		assert.NoError(t, uc.RecordEvent(ctx, event))
	})

	t.Run("repository failure is returned for retry", func(t *testing.T) {
		repo, _, uc := setup(t)
		repo.EXPECT().Create(ctx, gomock.Any()).Return(problem.ErrRepositoryFailure)
		assert.ErrorIs(t, uc.RecordEvent(ctx, event), problem.ErrRepositoryFailure)
	})

	t.Run("push failure does not fail a stored event", func(t *testing.T) {
		repo, notifier, uc := setup(t)
		repo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		notifier.EXPECT().Notify(ctx, gomock.Any()).Return(false, errors.New("stream closed"))
		assert.NoError(t, uc.RecordEvent(ctx, event))
	})

	t.Run("unsupported schema version is rejected before storing", func(t *testing.T) {
		_, _, uc := setup(t)
		unsupported := event
		unsupported.SchemaVersion = "2.0"
		assert.ErrorIs(t, uc.RecordEvent(ctx, unsupported), domain.ErrUnsupportedEventSchema)
	})
}

func TestNotificationUseCase_GetNotifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockActorLookup(ctrl)
	lgr := &mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mocks.NewMockRealtimeNotifier(ctrl), lgr)

	ctx := context.Background()
	username := "testuser"
	limit := 10

	user := &auth.Principal{
		ID:       1,
		Username: username,
	}

	t.Run("success with caching and ignored not found actor lookups", func(t *testing.T) {
		notifications := []*domain.Notification{
			{ID: 101, UserID: 1, ActorID: 2, Type: domain.NotificationTypeFollow},
			{ID: 102, UserID: 1, ActorID: 2, Type: domain.NotificationTypeLike},
			{ID: 103, UserID: 1, ActorID: 3, Type: domain.NotificationTypeComment},
			{ID: 104, UserID: 1, ActorID: 4, Type: domain.NotificationTypeComment},
			{ID: 105, UserID: 1, ActorID: 4, Type: domain.NotificationTypeComment}, // Second notification from the same missing actor ID 4
			nil, // Nil notification to test skipping without dereferencing/panicking
			{ID: 106, UserID: 1, ActorID: 1, Type: domain.NotificationTypeComment}, // Notification where actor is the current user (ID 1), should use pre-populated cache
		}
		mockRepo.EXPECT().GetByUserID(ctx, user.ID, nil, limit+1).Return(notifications, nil)

		// Actor 2: lookup succeeds once
		actor2 := &domain.Actor{ID: 2, Username: "actor2"}
		mockUserRepo.EXPECT().GetUserByID(ctx, 2).Return(actor2, nil).Times(1)

		// Actor 3: lookup succeeds
		actor3 := &domain.Actor{ID: 3, Username: "actor3"}
		mockUserRepo.EXPECT().GetUserByID(ctx, 3).Return(actor3, nil)

		// Actor 4: lookup fails with ErrUserNotFound, should only be called once due to caching of soft failure
		mockUserRepo.EXPECT().GetUserByID(ctx, 4).Return(nil, problem.ErrUserNotFound).Times(1)

		results, err := uc.GetNotifications(ctx, user, nil, limit)
		assert.NoError(t, err)
		assert.Len(t, results.Notifications, 6)

		assert.Equal(t, "actor2", results.Notifications[0].ActorUsername)
		assert.Equal(t, "actor2", results.Notifications[1].ActorUsername)
		assert.Equal(t, "actor3", results.Notifications[2].ActorUsername)
		assert.Equal(t, "", results.Notifications[3].ActorUsername)
		assert.Equal(t, "", results.Notifications[4].ActorUsername)
		assert.Equal(t, username, results.Notifications[5].ActorUsername)
	})

	t.Run("invalid recipient", func(t *testing.T) {
		results, err := uc.GetNotifications(ctx, nil, nil, limit)
		assert.ErrorIs(t, err, problem.ErrInvalidUser)
		assert.Nil(t, results)
	})

	t.Run("notification repo error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockRepo.EXPECT().GetByUserID(ctx, user.ID, nil, limit+1).Return(nil, expectedErr)

		results, err := uc.GetNotifications(ctx, user, nil, limit)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, results)
	})

	t.Run("actor lookup generic error", func(t *testing.T) {
		notifications := []*domain.Notification{
			{ID: 101, UserID: 1, ActorID: 5, Type: domain.NotificationTypeFollow},
		}
		mockRepo.EXPECT().GetByUserID(ctx, user.ID, nil, limit+1).Return(notifications, nil)

		expectedErr := errors.New("connection failed")
		mockUserRepo.EXPECT().GetUserByID(ctx, 5).Return(nil, expectedErr)

		results, err := uc.GetNotifications(ctx, user, nil, limit)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, results)
	})
}

func TestNotificationUseCase_MarkAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockActorLookup(ctrl)
	lgr := &mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mocks.NewMockRealtimeNotifier(ctrl), lgr)

	ctx := context.Background()
	username := "testuser"
	notificationID := 101

	user := &auth.Principal{
		ID:       1,
		Username: username,
	}

	t.Run("success", func(t *testing.T) {
		notification := &domain.Notification{ID: notificationID, UserID: user.ID}
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(notification, nil)
		mockRepo.EXPECT().MarkAsRead(ctx, notificationID).Return(nil)

		err := uc.MarkAsRead(ctx, notificationID, user.ID)
		assert.NoError(t, err)
	})

	t.Run("invalid user id", func(t *testing.T) {
		err := uc.MarkAsRead(ctx, notificationID, 0)
		assert.ErrorIs(t, err, problem.ErrInvalidUser)
	})

	t.Run("notification repo get error", func(t *testing.T) {
		expectedErr := errors.New("not found")
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(nil, expectedErr)

		err := uc.MarkAsRead(ctx, notificationID, user.ID)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("notification nil check", func(t *testing.T) {
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(nil, nil)

		err := uc.MarkAsRead(ctx, notificationID, user.ID)
		assert.ErrorIs(t, err, domain.ErrNotificationNotFound)
	})

	t.Run("unauthorized owner mismatch", func(t *testing.T) {
		notification := &domain.Notification{ID: notificationID, UserID: 999}
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(notification, nil)

		err := uc.MarkAsRead(ctx, notificationID, user.ID)
		assert.ErrorIs(t, err, domain.ErrNotificationNotOwned)
	})

	t.Run("mark as read repository error", func(t *testing.T) {
		expectedErr := errors.New("db write error")
		notification := &domain.Notification{ID: notificationID, UserID: user.ID}
		mockRepo.EXPECT().GetByID(ctx, notificationID).Return(notification, nil)
		mockRepo.EXPECT().MarkAsRead(ctx, notificationID).Return(expectedErr)

		err := uc.MarkAsRead(ctx, notificationID, user.ID)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestNotificationUseCase_MarkAllAsRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockActorLookup(ctrl)
	lgr := &mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mocks.NewMockRealtimeNotifier(ctrl), lgr)

	ctx := context.Background()
	username := "testuser"

	user := &auth.Principal{
		ID:       1,
		Username: username,
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().MarkAllAsRead(ctx, user.ID).Return(nil)

		err := uc.MarkAllAsRead(ctx, user.ID)
		assert.NoError(t, err)
	})

	t.Run("invalid user id", func(t *testing.T) {
		err := uc.MarkAllAsRead(ctx, 0)
		assert.ErrorIs(t, err, problem.ErrInvalidUser)
	})

	t.Run("mark all as read repository error", func(t *testing.T) {
		expectedErr := errors.New("db write error")
		mockRepo.EXPECT().MarkAllAsRead(ctx, user.ID).Return(expectedErr)

		err := uc.MarkAllAsRead(ctx, user.ID)
		assert.ErrorIs(t, err, expectedErr)
	})
}
