package core

import (
	"context"
	"encoding/json"
	"fmt"
	"hub/utils"

	"log"
	"os"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func TechSupport(nsecForHost string, npubForHost string, channelId string) {
	fmt.Println(npubForHost, "online")
	fmt.Println("channelId", channelId)
	fmt.Println()

	ctx := context.Background()

	nostr.InfoLogger = log.New(os.Stderr, "[go-nostr][info] ", log.LstdFlags)
	nostr.DebugLogger = log.New(os.Stderr, "[go-nostr][debug] ", log.LstdFlags)

	url := os.Getenv("HUB_RELAY")
	maxRetries := 5

	// Channel to signal when connectAndListen exits
	retrySignal := make(chan error)

	// Function to establish the connection and listen for events
	connectAndListen := func(ctx context.Context) {
		defer func() {
			retrySignal <- fmt.Errorf("connectAndListen terminated unexpectedly")
		}()

		// Connect to the relay
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			log.Printf("Failed to connect to relay: %v", err)
			retrySignal <- fmt.Errorf("failed to connect to relay: %v", err)
			return
		}
		defer relay.Close()

		fmt.Println("Listening for nostr events...")

		_, v1, _ := nip19.Decode(npubForHost)
		tags := map[string][]string{"p": {v1.(string)}}
		filters := []nostr.Filter{{
			Kinds: []int{nostr.KindEncryptedDirectMessage},
			Tags:  tags,
			Limit: 1,
		}}

		sub, err := relay.Subscribe(ctx, filters)
		if err != nil {
			log.Printf("Failed to subscribe to relay: %v", err)
			retrySignal <- fmt.Errorf("failed to subscribe to relay: %v", err)
			return
		}

		for {
			select {
			case event := <-sub.Events:
				// Process incoming events
				handleEvent(event, nsecForHost, npubForHost, channelId)
			case <-ctx.Done():
				log.Println("Context canceled, exiting connectAndListen...")
				return
			case <-relay.Context().Done():
				log.Println("Relay context done, closing connection...")
				return
			}
		}
	}

	// Retry loop
	retryCount := 0
	for retryCount < maxRetries {
		log.Printf("Attempting to start connectAndListen... (Retry %d/%d)", retryCount+1, maxRetries)
		go connectAndListen(ctx)

		// Wait for connectAndListen to exit
		err := <-retrySignal
		if err != nil {
			log.Printf("Error: %v. Retrying in 5 seconds... (Attempt %d/%d)", err, retryCount+1, maxRetries)
			time.Sleep(5 * time.Second)
			retryCount++
		} else {
			log.Println("Connection successfully established and running.")
			break
		}

		if retryCount == maxRetries {
			log.Println("Max retries reached. Could not reconnect to the relay.")
		}
	}

	fmt.Println("done")
}

// Handle individual events
func handleEvent(event *nostr.Event, nsecForHost, npubForHost, channelId string) {
	_, sk, _ := nip19.Decode(nsecForHost)
	shared, _ := nip04.ComputeSharedSecret(event.PubKey, sk.(string))
	npub, _ := nip19.EncodePublicKey(event.PubKey)

	ciphertext := event.Content
	plaintext, _ := nip04.Decrypt(ciphertext, shared)

	var message Message
	err := json.Unmarshal([]byte(plaintext), &message)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		fmt.Println(npub, ":", plaintext)
	} else {
		handleMessageContent(message, npub, nsecForHost, npubForHost, channelId)
	}
}

// Helper function to handle specific message content logic
func handleMessageContent(message Message, npub, nsecForHost, npubForHost, channelId string) {
	if message.Content == "🙂" {
		utils.PublishEncrypted(npub, "🙃")
	}

	if strings.Contains(message.Content, "Hi, I would like to report ") {
		reply := fmt.Sprintf(
			"Could you elaborate on the problem you're encountering with %s? Additional details would greatly assist in resolving your issue. In the meanwhile, feel free to mute the user if that's necessary.",
			utils.ExtractUsername(message.Content),
		)
		utils.PublishEncrypted(npub, reply)
	}

	if message.Content == "I'm online." {
		welcomeMessage := "Welcome to SkateConnect, skater! If you have any questions or need to report a bug do not hesitate to get in touch with us."
		utils.PublishEncrypted(npub, welcomeMessage)
		Announce(channelId, npub, nsecForHost, npubForHost)
	}
}
