package bots

import (
	"context"
	"encoding/json"
	"fmt"
	"hub/core"
	"log"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type DMBot struct {
	RelayURL         string
	SecretKey        string
	PublicKey        string
	Context          context.Context
	CancelFunc       context.CancelFunc
	Relay            *nostr.Relay
	IsActiveListener bool
	Handler          MessageHandler
	Plugins          []BotPlugin
}

// NewDMBot initializes a new instance of DMBot
func NewDMBot(relayURL, nsec string, handler MessageHandler, plugins []BotPlugin) *DMBot {
	_, sk, _ := nip19.Decode(nsec)
	pk, _ := nostr.GetPublicKey(sk.(string))
	npub, _ := nip19.EncodePublicKey(pk)

	ctx, cancel := context.WithCancel(context.Background())
	return &DMBot{
		RelayURL:         relayURL,
		SecretKey:        nsec,
		PublicKey:        npub,
		Context:          ctx,
		CancelFunc:       cancel,
		IsActiveListener: false,
		Handler:          handler,
		Plugins:          plugins,
	}
}

// Start begins listening for direct messages
func (bot *DMBot) Start() {
	retrySignal := make(chan error)
	for {
		select {
		case <-bot.Context.Done():
			log.Println("🛑 Bot shutting down...")
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

// Stop gracefully stops the bot
func (bot *DMBot) Stop() {
	bot.CancelFunc()
	if bot.Relay != nil {
		bot.Relay.Close()
	}
	log.Println("🛑 Bot stopped gracefully")
}

// IsReady checks if the bot is in the active listener state
func (bot *DMBot) IsReady() bool {
	return bot.IsActiveListener
}

// GetSecretKey returns the bot's secret key
func (bot *DMBot) GetSecretKey() string {
	return bot.SecretKey
}

// GetPublicKey returns the bot's public key
func (bot *DMBot) GetPublicKey() string {
	return bot.PublicKey
}

// PublishEncrypted sends an encrypted direct message
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

	bot.subscribeAndProcessEvents(relay)
}

func (bot *DMBot) subscribeAndProcessEvents(relay *nostr.Relay) {
	_, pubKeyDecoded, _ := nip19.Decode(bot.PublicKey)
	tags := map[string][]string{"p": {pubKeyDecoded.(string)}}
	filters := []nostr.Filter{{
		Kinds: []int{nostr.KindEncryptedDirectMessage},
		Tags:  tags,
		Limit: 50,
	}}

	sub, err := relay.Subscribe(bot.Context, filters)
	if err != nil {
		log.Printf("Subscription failed: %v", err)
		return
	}
	defer sub.Unsub()

	bot.processEvents(sub)
}

func (bot *DMBot) processEvents(sub *nostr.Subscription) {
	var storedEvents []*nostr.Event
	processingStoredEvents := false

	for {
		select {
		case event := <-sub.Events:
			if !processingStoredEvents {
				storedEvents = append(storedEvents, event)
			} else if bot.IsActiveListener {
				go bot.HandleEvent(event)
			}

		case <-sub.EndOfStoredEvents:
			if !processingStoredEvents {
				for _, event := range storedEvents {
					bot.HandleEvent(event)
				}
				storedEvents = nil
				processingStoredEvents = true
				bot.IsActiveListener = true
				log.Println("🚀 Bot is now in active listening mode")
			}

		case <-bot.Relay.Context().Done():
			log.Println("🚫 Relay connection lost, reconnecting...")
			return
		}
	}
}

// HandleEvent decrypts and processes an incoming event
func (bot *DMBot) HandleEvent(event *nostr.Event) {

	for _, plugin := range bot.Plugins {
		plugin.OnEvent(bot, event)
	}

	_, sk, _ := nip19.Decode(bot.SecretKey)
	shared, _ := nip04.ComputeSharedSecret(event.PubKey, sk.(string))
	npub, _ := nip19.EncodePublicKey(event.PubKey)

	plaintext, err := nip04.Decrypt(event.Content, shared)
	if err != nil {
		log.Printf("❌ Decryption failed: %v", err)
		return
	}

	var message core.Message
	if err := json.Unmarshal([]byte(plaintext), &message); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	if bot.IsActiveListener {
		bot.Handler.HandleMessage(bot, message, npub)
	}
}
