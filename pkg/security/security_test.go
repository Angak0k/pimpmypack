package security

import (
	"errors"
	"reflect"
	"testing"
)

func TestVerifyPassword(t *testing.T) {
	firsthashedPassword, err := HashPassword("password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	secondhashedPassword, err := HashPassword("password1")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	type args struct {
		password       string
		hashedPassword string
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "TestVerifyPassword",
			args: args{
				password:       "password",
				hashedPassword: firsthashedPassword,
			},
			want: nil,
		},
		{
			name: "TestVerifyPassword",
			args: args{
				password:       "password",
				hashedPassword: secondhashedPassword,
			},
			want: errors.New("crypto/bcrypt: hashedPassword is not the hash of the given password"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyPassword(tt.args.password, tt.args.hashedPassword); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
