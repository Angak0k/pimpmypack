package security

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyPassword(t *testing.T) {
	firstHashedPassword, err := HashPassword("password")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	secondHashedPassword, err := HashPassword("password1")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name string
		args struct {
			password       string
			hashedPassword string
		}
		wantErr     bool
		wantErrType error
	}{
		{
			name: "valid password",
			args: struct {
				password       string
				hashedPassword string
			}{
				password:       "password",
				hashedPassword: firstHashedPassword,
			},
			wantErr: false,
		},
		{
			name: "invalid password",
			args: struct {
				password       string
				hashedPassword string
			}{
				password:       "password",
				hashedPassword: secondHashedPassword,
			},
			wantErr:     true,
			wantErrType: bcrypt.ErrMismatchedHashAndPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.args.password, tt.args.hashedPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, tt.wantErrType) {
				t.Errorf("VerifyPassword() error = %v, wantErr type %v", err, tt.wantErrType)
			}
		})
	}
}
