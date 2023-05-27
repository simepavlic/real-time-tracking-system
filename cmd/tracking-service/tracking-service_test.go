package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func Test_populateAccounts(t *testing.T) {
	server := miniredis.RunT(t)
	redisClient = redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	tests := []struct {
		name     string
		accounts []Account
		wantErr  bool
	}{
		{
			name: "successful test",
			accounts: []Account{
				{
					ID:       "1",
					Name:     "Acc1",
					IsActive: true,
				},
				{
					ID:       "2",
					Name:     "Acc2",
					IsActive: false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := populateAccounts(tt.accounts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateAccount(t *testing.T) {
	// Simulate a test account
	accounts := []Account{
		{
			ID:       "1",
			Name:     "Acc1",
			IsActive: true,
		},
		{
			ID:       "2",
			Name:     "Acc2",
			IsActive: false,
		},
	}

	server := miniredis.RunT(t)
	redisClient = redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	err := populateAccounts(accounts)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		accountID string
		active    bool
		wantErr   bool
	}{
		{
			name:      "active account",
			accountID: "1",
			active:    true,
			wantErr:   false,
		},
		{
			name:      "inactive account",
			accountID: "2",
			active:    false,
			wantErr:   false,
		},
		{
			name:      "invalid account",
			accountID: "3",
			active:    false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isActive, err := validateAccount(tt.accountID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.active, isActive)
		})
	}
}

func TestEventHandler(t *testing.T) {
	server := miniredis.RunT(t)
	redisClient = redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	accounts := []Account{
		{
			ID:       "1",
			Name:     "Acc1",
			IsActive: true,
		},
		{
			ID:       "2",
			Name:     "Acc2",
			IsActive: false,
		},
	}
	err := populateAccounts(accounts)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		accountID  string
		httpStatus int
		body       string
	}{
		{
			name:       "active account",
			accountID:  "1",
			httpStatus: http.StatusOK,
			body:       "Event processed successfully",
		},
		{
			name:       "inactive account",
			accountID:  "2",
			httpStatus: http.StatusBadRequest,
			body:       "Account is not active\n",
		},
		{
			name:       "invalid account",
			accountID:  "3",
			httpStatus: http.StatusInternalServerError,
			body:       "Failed to validate account\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP request with a valid account ID and data
			req, err := http.NewRequest("GET", fmt.Sprintf("/%s?data=test", tt.accountID), nil)
			assert.NoError(t, err)

			// Create a response recorder for capturing the response
			rr := httptest.NewRecorder()

			// Call the eventHandler function
			eventHandler(rr, req)

			// Check the response status code
			assert.Equal(t, tt.httpStatus, rr.Code)
			assert.Equal(t, tt.body, rr.Body.String())
		})
	}
}
