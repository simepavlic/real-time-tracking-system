package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

type Account struct {
	ID       string `json:"accountId"`
	Name     string `json:"accountName"`
	IsActive bool   `json:"isActive"`
}

var (
	redisClient *redis.Client
)

func initRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Check the connectivity with Redis
	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
}

func populateAccounts(accounts []Account) error {

	// Populate Redis with account data
	for _, account := range accounts {
		accountData, err := json.Marshal(account)
		if err != nil {
			return fmt.Errorf("failed to marshal account '%s' data: %w", account.ID, err)
		}

		// Use HSet to store the account data in the Redis hash
		err = redisClient.HSet("accounts", account.ID, string(accountData)).Err()
		if err != nil {
			return fmt.Errorf("failed to populate Redis with account '%s': %w", account.ID, err)
		}
	}

	return nil
}

func validateAccount(accountID string) (bool, error) {
	// Query the database for the account's isActive field
	jsonResult, err := redisClient.HGet("accounts", accountID).Result()
	if err != nil {
		return false, err
	}
	var result Account
	err = json.Unmarshal([]byte(jsonResult), &result)
	if err != nil {
		return false, fmt.Errorf("unable to unmarshal the result from database: %w", err)
	}

	return result.IsActive, nil
}

type EventPayload struct {
	AccountID string `json:"accountId"`
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}

func propagateEvent(payload EventPayload) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to marshal event payload to JSON:", err)
		return
	}

	// Publish the JSON payload to Redis Pub/Sub channel
	err = redisClient.Publish("tracking-events", jsonPayload).Err()
	if err != nil {
		log.Println("Failed to publish event to Redis Pub/Sub:", err)
		return
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the account ID from the URL path
	accountID := r.URL.Path[1:]

	// Validate the account ID against the database
	isActive, err := validateAccount(accountID)
	if err != nil {
		log.Println("Failed to validate account:", err)
		http.Error(w, "Failed to validate account", http.StatusInternalServerError)
		return
	}

	if !isActive {
		log.Println("Account is not active:", accountID)
		http.Error(w, "Account is not active", http.StatusBadRequest)
		return
	}

	// Parse the data parameter from the query string
	data := r.URL.Query().Get("data")

	// Create an event payload
	payload := EventPayload{
		AccountID: accountID,
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	// Publish the event payload to Redis Pub/Sub
	go propagateEvent(payload)

	// Send a response indicating successful event processing
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event processed successfully"))
}

func main() {
	initRedisClient()
	accounts := []Account{
		{ID: "1", Name: "Account 1", IsActive: true},
		{ID: "2", Name: "Account 2", IsActive: false},
		{ID: "3", Name: "Account 3", IsActive: true},
		{ID: "4", Name: "Account 4", IsActive: true},
	}
	err := populateAccounts(accounts)
	if err != nil {
		log.Printf("failed to populate accounts to redis: %v", err)
		log.Fatal()
	}

	// Define the endpoint for receiving events
	http.HandleFunc("/", eventHandler)

	// Start the server on port 8080
	fmt.Println("Tracking service started. Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
