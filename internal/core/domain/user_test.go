package domain

import (
	"errors"
	"testing"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name: "valid user",
			user: &User{
				Username:  "johndoe",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: nil,
		},
		{
			name: "username too short",
			user: &User{
				Username:  "jo",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "username with spaces",
			user: &User{
				Username:  "john doe",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "username with special characters",
			user: &User{
				Username:  "john@doe",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: ErrValidationFailed,
		},
		{
			name: "username with underscore is valid",
			user: &User{
				Username:  "john_doe",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "password123",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}
