package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis"
)

var filterAccountIDs = []string{"1", "3"}

type EventPayload struct {
	AccountID string `json:"accountId"`
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}

func displayEvent(payload EventPayload) {
	fmt.Printf("Received event: %+v\n", payload)
}
func contains(id string, array []string) bool {
	for _, elem := range array {
		if id == elem {
			return true
		}
	}
	return false
}

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Check the connectivity with Redis
	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	pubsub := redisClient.Subscribe("tracking-events")
	defer pubsub.Close()

	// Wait for confirmation that subscription is created before subscribing to signals
	_, err = pubsub.Receive()
	if err != nil {
		log.Fatal("Failed to subscribe to Redis Pub/Sub channel:", err)
	}

	ch := pubsub.Channel()

	// Channel to receive termination signal from OS
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Map to store the last received timestamp per account ID
	lastTimestamps := make(map[string]int64)

	fmt.Println("CLI client started. Listening for events...")

	for {
		select {
		case msg := <-ch:
			var payload EventPayload
			err := json.Unmarshal([]byte(msg.Payload), &payload)
			if err != nil {
				log.Println("Failed to unmarshal event payload:", err)
				continue
			}

			// Check if the account ID matches the filter
			if !contains(payload.AccountID, filterAccountIDs) {
				continue
			}

			// Check if the event is newer than the last received event for the same account ID
			if lastTimestamp, ok := lastTimestamps[payload.AccountID]; ok {
				if payload.Timestamp <= lastTimestamp {
					continue
				}
			}

			// Update the last received timestamp for the account ID
			lastTimestamps[payload.AccountID] = payload.Timestamp

			// Display the event
			displayEvent(payload)

		case <-signalCh:
			// Termination signal received, clean up and exit
			fmt.Println("\nTerminating CLI client...")
			return
		}
	}
}
