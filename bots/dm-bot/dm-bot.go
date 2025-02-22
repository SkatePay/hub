package dmbot

import (
	"context"
	"encoding/json"
	"fmt"
	"hub/core"
	"log"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// DMBot encapsulates the state and behavior of the bot
type DMBot struct {
	SecretKey        string
	PublicKey        string
	RelayURL         string
	ChannelID        string
	Context          context.Context
	CancelFunc       context.CancelFunc
	Relay            *nostr.Relay
	IsActiveListener bool // Tracks if the bot is in the active listener state
}

// NewDMBot initializes the bot instance
func NewDMBot(nsec, npub, relayURL, channelID string) (*DMBot, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &DMBot{
		SecretKey:        nsec,
		PublicKey:        npub,
		RelayURL:         relayURL,
		ChannelID:        channelID,
		Context:          ctx,
		CancelFunc:       cancel,
		IsActiveListener: false,
	}, nil
}

// Start begins listening for direct messages
func (bot *DMBot) Start() {
	retrySignal := make(chan error)

	for {
		select {
		case <-bot.Context.Done():
			log.Println("🛑 Bot is shutting down gracefully...")
			return
		default:
			go bot.connectAndListen(retrySignal)

			if err := <-retrySignal; err != nil {
				log.Printf("❌ Error: %v. Retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// connectAndListen establishes a connection and subscribes to events
func (bot *DMBot) connectAndListen(retrySignal chan error) {
	defer func() {
		retrySignal <- fmt.Errorf("connectAndListen terminated unexpectedly")
	}()

	relay, err := nostr.RelayConnect(bot.Context, bot.RelayURL)
	if err != nil {
		log.Printf("Failed to connect to relay: %v", err)
		retrySignal <- err
		return
	}
	bot.Relay = relay
	defer relay.Close()

	log.Println("✅ Connected to relay and fetching pending events...")

	_, pubKeyDecoded, _ := nip19.Decode(bot.PublicKey)
	tags := map[string][]string{"p": {pubKeyDecoded.(string)}}
	filters := []nostr.Filter{{
		Kinds: []int{nostr.KindEncryptedDirectMessage},
		Tags:  tags,
		Limit: 50,
	}}

	sub, err := relay.Subscribe(bot.Context, filters)
	if err != nil {
		log.Printf("Failed to subscribe to relay: %v", err)
		retrySignal <- err
		return
	}
	defer sub.Unsub()

	bot.processEvents(sub)

	for {
		select {
		case <-bot.Context.Done():
			log.Println("Context canceled, stopping event listener...")
			return
		case <-relay.Context().Done():
			log.Println("Relay connection lost, reconnecting...")
			return
		}
	}
}

// processEvents handles pending events and switches to active listener state
func (bot *DMBot) processEvents(sub *nostr.Subscription) {
	var storedEvents []*nostr.Event
	processingStoredEvents := false

	for {
		select {
		case event := <-sub.Events:
			if !processingStoredEvents {
				storedEvents = append(storedEvents, event)
			} else {
				if bot.IsActiveListener {
					go bot.handleEvent(event)
				}
			}

		case <-sub.EndOfStoredEvents:
			if !processingStoredEvents {
				log.Println("📥 Processing pending events...")
				for i := len(storedEvents) - 1; i >= 0; i-- {
					event := storedEvents[i]
					bot.handleEvent(event)
				}
				storedEvents = nil
				processingStoredEvents = true
				bot.IsActiveListener = true
				log.Println("🚀 Bot is now in active listening mode")
			}
		}
	}
}

// handleEvent decrypts and processes an incoming event
func (bot *DMBot) handleEvent(event *nostr.Event) {
	_, sk, _ := nip19.Decode(bot.SecretKey)
	shared, _ := nip04.ComputeSharedSecret(event.PubKey, sk.(string))
	npub, _ := nip19.EncodePublicKey(event.PubKey)

	plaintext, err := nip04.Decrypt(event.Content, shared)
	if err != nil {
		log.Printf("❌ Failed to decrypt message: %v", err)
		return
	}

	var message core.Message
	if err := json.Unmarshal([]byte(plaintext), &message); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		fmt.Println(npub, ":", plaintext)
	} else {
		bot.handleMessageContent(message, npub)
	}
}

// handleMessageContent processes the decrypted message and sends appropriate replies
func (bot *DMBot) handleMessageContent(message core.Message, npub string) {
	if !bot.IsActiveListener {
		log.Println("⚠️ Incoming message ignored: Bot is still processing pending events")
		return
	}

	switch {
	case message.Content == "🙂":
		bot.PublishEncrypted(npub, "🙃")

	case strings.Contains(message.Content, "Hi, I would like to report "):
		reply := fmt.Sprintf(
			"Could you elaborate on the problem you're encountering with %s? Additional details would greatly assist in resolving your issue. In the meanwhile, feel free to mute the user if that's necessary.",
			bot.ExtractUsername(message.Content),
		)
		bot.PublishEncrypted(npub, reply)

	case message.Content == "I'm online.":
		welcomeMessage := "Welcome to SkateConnect! If you have any questions or need help, feel free to ask."
		bot.PublishEncrypted(npub, welcomeMessage)
		core.Announce(bot.ChannelID, npub, bot.SecretKey, bot.PublicKey)
	}
}

// ExtractUsername extracts a username from a message
func (bot *DMBot) ExtractUsername(input string) string {
	input = strings.TrimSuffix(input, ".")
	length := len(input)
	if length > 10 {
		return input[length-10:]
	}
	return input
}

// PublishEncrypted encrypts and sends a message
func (bot *DMBot) PublishEncrypted(npubReceiver, message string) {
	_, receiverKey, err := nip19.Decode(npubReceiver)
	if err != nil {
		log.Printf("❌ Failed to decode receiver's public key: %v", err)
		return
	}

	_, sk, err := nip19.Decode(bot.SecretKey)
	if err != nil {
		log.Printf("❌ Error decoding private key: %v", err)
		return
	}

	shared, _ := nip04.ComputeSharedSecret(receiverKey.(string), sk.(string))
	encryptedMessage, _ := nip04.Encrypt(message, shared)

	ev := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindEncryptedDirectMessage,
		Tags:      nostr.Tags{{"p", receiverKey.(string)}},
		Content:   encryptedMessage,
	}

	ev.Sign(sk.(string))

	if err := bot.Relay.Publish(bot.Context, ev); err != nil {
		log.Printf("❌ Failed to publish encrypted message: %v", err)
	} else {
		log.Printf("✉️ ➡️ Sent encrypted message to %s", npubReceiver)
	}
}

// Stop gracefully stops the bot and closes the relay connection
func (bot *DMBot) Stop() {
	bot.CancelFunc()
	if bot.Relay != nil {
		bot.Relay.Close()
	}
	log.Println("🛑 Bot stopped gracefully")
}
