package main

import (
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
			name: "basic test",
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

	populateAccounts(accounts)

	// Test an active account
	isActive, err := validateAccount("1")
	assert.NoError(t, err)
	assert.True(t, isActive)

	// Test an inactive account
	isActive, err = validateAccount("2")
	assert.NoError(t, err)
	assert.False(t, isActive)

	// Test an invalid account
	isActive, err = validateAccount("3")
	assert.Error(t, err)
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
	}
	err := populateAccounts(accounts)
	assert.NoError(t, err)

	// Create a mock HTTP request with a valid account ID and data
	req, err := http.NewRequest("GET", "/1?data=test", nil)
	assert.NoError(t, err)

	// Create a response recorder for capturing the response
	rr := httptest.NewRecorder()

	// Call the eventHandler function
	eventHandler(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Event processed successfully", rr.Body.String())
}
