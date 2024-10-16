package subscriber

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

const (
	maxRetries     = 5
	reconnectDelay = 5 * time.Second // Delay between reconnection attempts
)

func ChatBot(nsec string, npub string, channelId string) {
	url := os.Getenv("HUB_RELAY")

	_, pk, _ := nip19.Decode(npub)
	_, sk, _ := nip19.Decode(nsec)

	ctx := context.Background()

	// Set up logging for go-nostr
	nostr.InfoLogger = log.New(os.Stderr, "[go-nostr][info] ", log.LstdFlags)
	nostr.DebugLogger = log.New(os.Stderr, "[go-nostr][debug] ", log.LstdFlags)

	// Function to establish the connection and subscribe to the relay
	connectAndSubscribe := func() (*nostr.Subscription, *nostr.Relay, error) {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			log.Printf("Failed to connect to relay: %v", err)
			return nil, nil, fmt.Errorf("failed to connect to relay: %v", err)
		}

		fmt.Println("📡", channelId, " connected")
		fmt.Println("🇺🇸", npub, "online")
		fmt.Println()

		tags := make(map[string][]string)
		tags["e"] = []string{channelId}

		filters := []nostr.Filter{{
			Kinds: []int{nostr.KindChannelMessage},
			Tags:  tags,
			Limit: 64,
		}}

		sub, err := relay.Subscribe(ctx, filters)
		if err != nil {
			log.Printf("Failed to subscribe to relay: %v", err)
			return nil, nil, fmt.Errorf("failed to subscribe to relay: %v", err)
		}

		return sub, relay, nil
	}

	// Reconnection logic with retries
	for retryCount := 0; retryCount < maxRetries; retryCount++ {
		sub, relay, err := connectAndSubscribe()
		if err != nil {
			fmt.Printf("Error connecting to relay: %v. Retrying in %v... (Attempt %d/%d)\n", err, reconnectDelay, retryCount+1, maxRetries)
			time.Sleep(reconnectDelay)
			continue
		}

		// Process events while connected
		err = processEvents(ctx, sub, relay, pk.(string), sk.(string), channelId)
		if err != nil {
			log.Printf("Error: %v. Retrying in %v... (Attempt %d/%d)", err, reconnectDelay, retryCount+1, maxRetries)
			time.Sleep(reconnectDelay)
			continue
		}

		// If processing completes without errors, break the retry loop
		break
	}

	fmt.Println("Max retries reached. Could not reconnect to the relay.")
}
