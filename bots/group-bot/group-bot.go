package groupbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"hub/core"
	"hub/services/weather"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// GroupBot manages the group message processing state and behavior
type GroupBot struct {
	SecretKey        string
	PublicKey        string
	RelayURL         string
	ChannelID        string
	Context          context.Context
	CancelFunc       context.CancelFunc
	Relay            *nostr.Relay
	IsActiveListener bool // Tracks if the bot is in active listening state
}

// NewGroupBot initializes a new instance of GroupBot
func NewGroupBot(nsec, npub, relayURL, channelID string) (*GroupBot, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &GroupBot{
		SecretKey:        nsec,
		PublicKey:        npub,
		RelayURL:         relayURL,
		ChannelID:        channelID,
		Context:          ctx,
		CancelFunc:       cancel,
		IsActiveListener: false,
	}, nil
}

// Start begins the connection and message subscription process
func (bot *GroupBot) Start() {
	const maxRetries = 5
	retryCount := 0

	for {
		select {
		case <-bot.Context.Done():
			log.Println("🛑 Bot shutting down gracefully...")
			return
		default:
			if err := bot.connectAndSubscribe(); err != nil {
				log.Printf("❌ Error: %v. Retrying in 5 seconds... (Attempt %d/%d)", err, retryCount+1, maxRetries)
				retryCount++
				if retryCount >= maxRetries {
					log.Println("❌ Max retries reached. Shutting down...")
					return
				}
				time.Sleep(5 * time.Second)
			} else {
				retryCount = 0
			}
		}
	}
}

// connectAndSubscribe establishes connection and processes events
func (bot *GroupBot) connectAndSubscribe() error {
	relay, err := nostr.RelayConnect(bot.Context, bot.RelayURL)
	if err != nil {
		return fmt.Errorf("failed to connect to relay: %v", err)
	}
	bot.Relay = relay
	defer relay.Close()

	log.Printf("✅ Connected to relay at %s\n", bot.RelayURL)

	// Subscribe to channel messages
	filters := []nostr.Filter{{
		Kinds: []int{nostr.KindChannelMessage},
		Tags:  map[string][]string{"e": {bot.ChannelID}},
		Limit: 64,
	}}

	sub, err := relay.Subscribe(bot.Context, filters)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %v", err)
	}
	defer sub.Unsub()

	return bot.processEvents(sub)
}

// processEvents handles the relay events (pending and active)
func (bot *GroupBot) processEvents(sub *nostr.Subscription) error {
	var storedEvents []*nostr.Event
	processingStoredEvents := false

	for {
		select {
		case event, ok := <-sub.Events:
			if !ok {
				log.Println("🚫 Subscription closed, reconnecting...")
				bot.Relay.Close()
				return fmt.Errorf("subscription closed")
			}

			if !processingStoredEvents {
				storedEvents = append(storedEvents, event)
			} else if bot.IsActiveListener {
				go bot.handleEvent(event)
			}

		case <-sub.EndOfStoredEvents:
			if !processingStoredEvents {
				log.Println("📥 Processing pending events...")
				for i := len(storedEvents) - 1; i >= 0; i-- {
					bot.handleEvent(storedEvents[i])
				}
				storedEvents = nil
				processingStoredEvents = true
				bot.IsActiveListener = true
				log.Println("🚀 Entered active listening mode")
			}

		case <-bot.Relay.Context().Done():
			log.Println("🚫 Relay connection lost")
			bot.Relay.Close()
			return fmt.Errorf("relay connection lost")
		}
	}
}

// handleEvent processes individual events but only replies if in active listener mode
func (bot *GroupBot) handleEvent(event *nostr.Event) {
	npub, _ := nip19.EncodePublicKey(event.PubKey)
	suffix := npub[len(npub)-3:]
	username := fmt.Sprintf("skate-%s", suffix)

	var message core.Message
	if err := json.Unmarshal([]byte(event.Content), &message); err != nil {
		log.Printf("💬 [%s]: %s\n", username, event.Content)
		return
	}

	log.Printf("💬 [%s]: %s\n", username, message.Content)

	// Only handle commands if the bot is in active listener mode
	if bot.IsActiveListener {
		bot.handleCommand(message, username)
	} else {
		// Log the pending event without replying
		log.Printf("⏳ Ignoring pending message from [%s] during startup phase", username)
	}
}

// handleCommand processes messages and sends appropriate replies
func (bot *GroupBot) handleCommand(message core.Message, username string) {
	switch {
	case strings.Contains(message.Content, "!weather"):
		bot.sendWeatherUpdate(username)
	}
}

// sendWeatherUpdate sends the weather report to the channel
func (bot *GroupBot) sendWeatherUpdate(username string) {
	content := weather.GetReport()

	event := nostr.Event{
		PubKey:    bot.PublicKey,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindChannelMessage,
		Content:   content,
		Tags: nostr.Tags{
			{"e", bot.ChannelID, bot.Relay.URL, "root"},
		},
	}

	_, sk, _ := nip19.Decode(bot.SecretKey)
	event.Sign(sk.(string))

	if err := bot.Relay.Publish(bot.Context, event); err != nil {
		log.Printf("❌ Failed to publish weather report: %v", err)
	} else {
		log.Printf("🌦️ Sent weather report to channel with %s", username)
	}
}

// Stop gracefully stops the bot and closes the relay
func (bot *GroupBot) Stop() {
	bot.CancelFunc()
	if bot.Relay != nil {
		bot.Relay.Close()
	}
	log.Println("🛑 Bot stopped gracefully")
}
