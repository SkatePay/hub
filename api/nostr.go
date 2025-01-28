package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

type RelayHelper struct {
	url        string
	ctx        context.Context
	maxRetries int
	delay      time.Duration
}

func NewRelayHelper(url string, maxRetries int, delay time.Duration) *RelayHelper {
	return &RelayHelper{
		url:        url,
		ctx:        context.Background(),
		maxRetries: maxRetries,
		delay:      delay,
	}
}

// Connect to the relay and return the subscription
func (r *RelayHelper) ConnectAndSubscribe(filters []nostr.Filter) (*nostr.Subscription, *nostr.Relay, error) {
	var relay *nostr.Relay
	var sub *nostr.Subscription
	var err error

	// Retry logic for reconnection
	for retryCount := 0; retryCount < r.maxRetries; retryCount++ {
		relay, err = nostr.RelayConnect(r.ctx, r.url)
		if err != nil {
			log.Printf("Failed to connect to relay: %v. Retrying in %v... (Attempt %d/%d)\n", err, r.delay, retryCount+1, r.maxRetries)
			time.Sleep(r.delay)
			continue
		}

		// Subscribe to the relay with the given filters
		sub, err = relay.Subscribe(r.ctx, filters)
		if err != nil {
			log.Printf("Failed to subscribe to relay: %v", err)
			time.Sleep(r.delay)
			continue
		}

		// Successful connection and subscription
		return sub, relay, nil
	}

	// Max retries reached
	return nil, nil, fmt.Errorf("max retries reached. could not connect to the relay")
}

func FetchEvents(nsec string, npub string, channelId string, tags map[string][]string) ([]nostr.Event, error) {
	url := os.Getenv("HUB_RELAY")

	helper := NewRelayHelper(url, 5, 5*time.Second) // RelayHelper with retry logic

	// Define filters
	filters := []nostr.Filter{{
		Kinds: []int{nostr.KindChannelMessage},
		Tags:  tags,
		Limit: 64,
	}}

	sub, relay, err := helper.ConnectAndSubscribe(filters)
	if err != nil {
		return nil, err
	}

	var events []nostr.Event
	processingStoredEvents := true

	for {
		select {
		case event := <-sub.Events:
			// Dereference and append the event
			events = append(events, *event)

		case <-sub.EndOfStoredEvents:
			// All cached events have been received, stop processing
			processingStoredEvents = false
			fmt.Println("End of stored events received, closing connection...")
			relay.Close()
			return events, nil

		case <-relay.Context().Done():
			// Relay context is done, meaning the connection is closed
			fmt.Println("Relay context done, closing connection...")
			relay.Close()
			return events, nil
		}

		if !processingStoredEvents {
			break
		}
	}

	// Return fetched events (in case no EndOfStoredEvents signal is received)
	return events, nil
}

// handleChannelVideos handles HTTP requests to fetch videos for a specific channel
func handleChannelVideos(w http.ResponseWriter, r *http.Request) {
	// Extract channelId from the URL path (e.g., /channel/{channelId})
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	channelId := parts[2]

	// Get the nsec and npub values from environment variables or configuration
	nsec := os.Getenv("HUB_NSEC")
	npub := os.Getenv("HUB_NPUB")

	// Define the tags for video events
	tags := make(map[string][]string)
	tags["e"] = []string{channelId}
	// tags["t"] = []string{"video"}

	// Fetch video events using the generic FetchEvents function
	videoEvents, err := FetchEvents(nsec, npub, channelId, tags)
	if err != nil {
		log.Printf("Failed to fetch videos for channelId=%s: %v", channelId, err)
		http.Error(w, "Failed to retrieve videos", http.StatusInternalServerError)
		return
	}

	// Convert the video events to JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]nostr.Event{
		"videos": videoEvents,
	})
}
