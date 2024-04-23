package config

import (
	"os"
	"strings"
	"testing"
)

func TestEnvInit(t *testing.T) {
	t.Run("Importing valid env file", func(t *testing.T) {
		// Test a valid .env file
		err := EnvInit("test/.env.testSuccess")
		if err != nil {
			t.Errorf("EnvInit failed: %v", err)
		}
	})

	t.Run("Importing invalid env file", func(t *testing.T) {
		// Test an invalid .env file
		err := EnvInit("test/.env.testFailure")
		if err == nil {
			t.Errorf("Expected an error, got nil")
		}
	})

	testCases := []struct {
		name      string   // Name of the subtest
		envSlice  []string // The .env file to test
		wantError bool     // Whether an error is expected
	}{
		{
			name: "Valid Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=localhost",
				"DB_USER=db_user",
				"DB_PASSWORD=db_password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=identity@exemple.com",
				"MAIL_USERNAME=username",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: false,
		},
		{
			name: "Invalid Hostname Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=",
				"DB_HOST=localhost",
				"DB_USER=db_user",
				"DB_PASSWORD=db_password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=identity@exemple.com",
				"MAIL_USERNAME=username",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: false,
		},
		{
			name: "Invalid DB_HOST Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=",
				"DB_USER=db_user",
				"DB_PASSWORD=db_password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=identity@exemple.com",
				"MAIL_USERNAME=username",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid DB_USER Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=",
				"DB_PASSWORD=db_password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=identity@exemple.com",
				"MAIL_USERNAME=username",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid DB_PASSWORD Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=db_user",
				"DB_PASSWORD=",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=identity@exemple.com",
				"MAIL_USERNAME=username",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid DB_NAME Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=db_user",
				"DB_PASSWORD=password",
				"DB_NAME=",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=identity@exemple.com",
				"MAIL_USERNAME=username",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid Mail Identity Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=db_user",
				"DB_PASSWORD=password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=",
				"MAIL_USERNAME=username",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid Mail user Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=db_user",
				"DB_PASSWORD=password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=pmp@exemple.com",
				"MAIL_USERNAME=",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid Mail password Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=db_user",
				"DB_PASSWORD=password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=pmp@exemple.com",
				"MAIL_USERNAME=user",
				"MAIL_PASSWORD=",
				"MAIL_SERVER=smtp.exemple.com",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid Mail server Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=db_user",
				"DB_PASSWORD=password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=pmp@exemple.com",
				"MAIL_USERNAME=user",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
		{
			name: "Invalid Mail server Configuration",
			envSlice: []string{
				"SCHEME=http",
				"HOSTNAME=localhost",
				"DB_HOST=hostname",
				"DB_USER=db_user",
				"DB_PASSWORD=password",
				"DB_NAME=db_name",
				"DB_PORT=5432",
				"STAGE=dev",
				"API_SECRET=API_SECRET",
				"TOKEN_HOUR_LIFESPAN=1",
				"MAIL_IDENTITY=pmp@exemple.com",
				"MAIL_USERNAME=user",
				"MAIL_PASSWORD=password",
				"MAIL_SERVER=",
				"MAIL_PORT=587",
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load environment varialble from test case
			for _, conf := range tc.envSlice {
				key := strings.Split(conf, "=")[0]
				value := strings.Split(conf, "=")[1]
				err := os.Setenv(key, value)
				if err != nil {
					t.Errorf("Error setting environment variable: %v", err)
				}
			}

			// Test EnvInit
			err := EnvInit("")
			if tc.wantError && err == nil {
				t.Errorf("Expected an error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
