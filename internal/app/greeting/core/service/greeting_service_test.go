package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/billykore/project-one/internal/app/greeting/core/domain"
)

// mockGreetingRepository is a manual mock implementation of GreetingRepository.
type mockGreetingRepository struct {
	getGreetingFunc func(ctx context.Context) (*domain.Greeting, error)
}

func (m *mockGreetingRepository) GetGreeting(ctx context.Context) (*domain.Greeting, error) {
	return m.getGreetingFunc(ctx)
}

func TestGreetingService_GetGreeting(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	tests := []struct {
		name    string
		mock    func(ctx context.Context) (*domain.Greeting, error)
		want    string
		wantErr bool
	}{
		{
			name: "successful retrieval",
			mock: func(ctx context.Context) (*domain.Greeting, error) {
				return &domain.Greeting{Message: "Hello, World!"}, nil
			},
			want:    "Hello, World!",
			wantErr: false,
		},
		{
			name: "repository error",
			mock: func(ctx context.Context) (*domain.Greeting, error) {
				return nil, errors.New("database error")
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid greeting",
			mock: func(ctx context.Context) (*domain.Greeting, error) {
				return &domain.Greeting{Message: ""}, nil
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockGreetingRepository{getGreetingFunc: tt.mock}
			svc := NewGreetingService(log, repo)

			got, err := svc.GetGreeting(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGreeting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Message != tt.want {
				t.Errorf("GetGreeting() got = %v, want %v", got.Message, tt.want)
			}
		})
	}
}
