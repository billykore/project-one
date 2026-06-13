package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/usecase/mocks"
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

func TestNotificationUseCase_SaveNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNotificationRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockConsumer := mocks.NewMockNotificationConsumer(ctrl)
	lgr := mockUseCaseLogger{}
	uc := NewNotificationUseCase(mockRepo, mockUserRepo, mockConsumer, lgr)

	ctx := context.Background()
	notification := &domain.Notification{ID: 101, UserID: 1, Type: domain.NotificationTypeFollow}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().Create(ctx, notification).Return(nil)
		err := uc.SaveNotification(ctx, notification)
		assert.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().Create(ctx, notification).Return(errors.New("db error"))
		err := uc.SaveNotification(ctx, notification)
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
