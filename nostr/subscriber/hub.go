package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"hub/nostr/publisher"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func getUsername(input string) string {
	input = strings.TrimSuffix(input, ".")
	length := len(input)
	return input[length-10:]
}

func TechSupport(nsecForHost string, npubForHost string, channelId string) {
	fmt.Println(npubForHost, "online")
	fmt.Println("channelId", channelId)
	fmt.Println()

	ctx := context.Background()

	nostr.InfoLogger = log.New(os.Stderr, "[go-nostr][info] ", log.LstdFlags)
	nostr.DebugLogger = log.New(os.Stderr, "[go-nostr][debug] ", log.LstdFlags)

	// Function to establish the connection and listen for events
	connectAndListen := func() error {
		// Connect to the relay
		relay, err := nostr.RelayConnect(ctx, "wss://relay.primal.net")
		if err != nil {
			log.Printf("Failed to connect to relay: %v", err)
			return fmt.Errorf("failed to connect to relay: %v", err)
		}
		defer relay.Close()

		fmt.Println("Listening for nostr events...")

		_, v1, _ := nip19.Decode(npubForHost)

		tags := make(map[string][]string)
		tags["p"] = []string{v1.(string)}

		filters := []nostr.Filter{{
			Kinds: []int{nostr.KindEncryptedDirectMessage},
			Tags:  tags,
			Limit: 1,
		}}

		// Subscribe to the relay
		sub, err := relay.Subscribe(ctx, filters)
		if err != nil {
			log.Printf("Failed to subscribe to relay: %v", err)
			return fmt.Errorf("failed to subscribe to relay: %v", err)
		}

		_, sk, _ := nip19.Decode(nsecForHost)

		// Event processing loop
		for {
			select {
			case event := <-sub.Events:
				// Process the events
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
					// Handle specific message content
					handleMessageContent(message, npub, nsecForHost, npubForHost, channelId)
				}

			case <-relay.Context().Done():
				// Relay context is done, this means the connection was lost
				log.Println("Relay context done, closing connection...")
				return fmt.Errorf("relay connection lost")
			}
		}
	}

	// Reconnection logic with retries and proper logging
	for retryCount := 0; retryCount < maxRetries; retryCount++ {
		err := connectAndListen()
		if err != nil {
			log.Printf("Error: %v. Retrying in 5 seconds... (Attempt %d/%d)", err, retryCount+1, maxRetries)
			time.Sleep(5 * time.Second)
		} else {
			// Successful connection, break out of retry loop
			// If the connection and event processing are successful, reset retryCount
			fmt.Println("Connection established and processed events successfully, resetting retry count.")
			retryCount = 0
			break
		}

		if retryCount == maxRetries-1 {
			log.Println("Max retries reached. Could not reconnect to the relay.")
		}
	}

	fmt.Println("done")
}

// Helper function to handle the specific message content logic
func handleMessageContent(message Message, npub, nsecForHost, npubForHost, channelId string) {
	if message.Content == "🙂" {
		publisher.Publish_Encrypted(npub, "🙃")
	}

	if strings.Contains(message.Content, "Hi, I would like to report ") {
		reply := fmt.Sprintf("Could you elaborate on the problem you're encountering with %s? Additional details would greatly assist in resolving your issue. In the meanwhile, feel free to mute the user if that's necessary.", getUsername(message.Content))
		publisher.Publish_Encrypted(npub, reply)
	}

	if message.Content == "I'm online." {
		welcomeMessage := "Welcome to SkateConnect, skater! If you have any questions or need to report a bug do not hesitate to get in touch with us."
		publisher.Publish_Encrypted(npub, welcomeMessage)
		Announce(channelId, npub, nsecForHost, npubForHost)
	}
}
