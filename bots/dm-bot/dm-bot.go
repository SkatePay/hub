package dmbot

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

// DMBot handles direct messages and tech support operations
type DMBot struct {
	SecretKey string
	PublicKey string
	RelayURL  string
	Context   context.Context
}

// NewDMBot initializes a new instance of DMBot
func NewDMBot(nsec, npub string) (*DMBot, error) {
	relayURL := os.Getenv("HUB_RELAY")
	if relayURL == "" {
		return nil, fmt.Errorf("HUB_RELAY environment variable is not set")
	}

	_, pk, err := nip19.Decode(npub)
	if err != nil {
		return nil, fmt.Errorf("failed to decode npub: %v", err)
	}
	_, sk, err := nip19.Decode(nsec)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nsec: %v", err)
	}

	return &DMBot{
		SecretKey: sk.(string),
		PublicKey: pk.(string),
		RelayURL:  relayURL,
		Context:   context.Background(),
	}, nil
}

// Start begins listening for direct messages
func (bot *DMBot) Start(channelID string) {
	// Set up logging for go-nostr
	nostr.InfoLogger = log.New(os.Stderr, "[dm-bot][info] ", log.LstdFlags)
	nostr.DebugLogger = log.New(os.Stderr, "[dm-bot][debug] ", log.LstdFlags)

	for retryCount := 0; retryCount < maxRetries; retryCount++ {
		sub, relay, err := bot.connectAndSubscribe(channelID)
		if err != nil {
			fmt.Printf("Error connecting to relay: %v. Retrying in %v... (Attempt %d/%d)\n", err, reconnectDelay, retryCount+1, maxRetries)
			time.Sleep(reconnectDelay)
			continue
		}

		// Process events while connected
		err = bot.processDirectMessages(sub, relay)
		if err != nil {
			log.Printf("Error: %v. Retrying in %v... (Attempt %d/%d)", err, reconnectDelay, retryCount+1, maxRetries)
			time.Sleep(reconnectDelay)
			continue
		}

		// Reset retry count on success
		fmt.Println("Connection established and processed events successfully, resetting retry count.")
		retryCount = 0
	}

	fmt.Println("Max retries reached. Could not reconnect to the relay.")
}

// connectAndSubscribe connects to the relay and subscribes to direct message events
func (bot *DMBot) connectAndSubscribe(channelID string) (*nostr.Subscription, *nostr.Relay, error) {
	relay, err := nostr.RelayConnect(bot.Context, bot.RelayURL)
	if err != nil {
		log.Printf("Failed to connect to relay: %v", err)
		return nil, nil, fmt.Errorf("failed to connect to relay: %v", err)
	}

	fmt.Println("📡 Connected to DM channel:", channelID)
	fmt.Println("🔑 Public key online:", bot.PublicKey)

	tags := map[string][]string{
		"p": {bot.PublicKey}, // Assuming direct messages use the recipient's pubkey tag
	}

	filters := []nostr.Filter{{
		Kinds: []int{nostr.KindEncryptedDirectMessage},
		Tags:  tags,
		Limit: 64,
	}}

	sub, err := relay.Subscribe(bot.Context, filters)
	if err != nil {
		log.Printf("Failed to subscribe to relay: %v", err)
		return nil, nil, fmt.Errorf("failed to subscribe to relay: %v", err)
	}

	return sub, relay, nil
}

// processDirectMessages handles incoming direct messages from the subscription
func (bot *DMBot) processDirectMessages(sub *nostr.Subscription, relay *nostr.Relay) error {
	for evt := range sub.Events {
		fmt.Printf("📥 Received DM: %v\n", evt.Content)
		// Decrypt and handle the DM logic here
	}
	return nil
}
