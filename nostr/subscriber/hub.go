package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"hub/nostr/publisher"
	"log"
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

	// Function to establish the connection and listen for events
	connectAndListen := func() error {
		relay, err := nostr.RelayConnect(ctx, "wss://relay.primal.net")
		if err != nil {
			return fmt.Errorf("failed to connect to relay: %v", err)
		}

		fmt.Println("Listening for nostr events...")

		_, v1, _ := nip19.Decode(npubForHost)

		tags := make(map[string][]string)
		tags["p"] = []string{v1.(string)}

		filters := []nostr.Filter{{
			Kinds: []int{nostr.KindEncryptedDirectMessage},
			Tags:  tags,
			Limit: 1,
		}}

		sub, err := relay.Subscribe(ctx, filters)
		if err != nil {
			return fmt.Errorf("failed to subscribe to relay: %v", err)
		}

		_, sk, _ := nip19.Decode(nsecForHost)

		for event := range sub.Events {
			shared, _ := nip04.ComputeSharedSecret(event.PubKey, sk.(string))

			npub, _ := nip19.EncodePublicKey(event.PubKey)
			fmt.Println()

			ciphertext := event.Content
			plaintext, _ := nip04.Decrypt(ciphertext, shared)

			var message Message
			err := json.Unmarshal([]byte(plaintext), &message)
			if err != nil {
				fmt.Print(err)
				fmt.Println(npub, ":", plaintext)
			} else {
				if message.Content == "🙂" {
					publisher.Publish_Encrypted(npub, "🙃")
				}

				if strings.Contains(message.Content, "Hi, I would like to report ") {
					message := fmt.Sprintf("Could you elaborate on the problem you're encountering with %s? Additional details would greatly assist in resolving your issue. In the meanwhile, feel free to mute the user if that's necessary.", getUsername(plaintext))
					publisher.Publish_Encrypted(npub, message)
				}

				if message.Content == "I’m online." {
					message := "Welcome to SkateConnect, skater! If you have any questions or need to report a bug do not hesitate to get in touch with us."
					publisher.Publish_Encrypted(npub, message)
					// Tell everyone who just joined Public Chat
					Announce(channelId, npub, nsecForHost, npubForHost)
				}
			}
		}

		fmt.Println("Subscription closed")
		return nil
	}

	// Reconnection logic with retries
	for retryCount := 0; retryCount < maxRetries; retryCount++ {
		err := connectAndListen()
		if err != nil {
			log.Printf("Error: %v. Retrying in 5 seconds... (Attempt %d/%d)", err, retryCount+1, maxRetries)
			time.Sleep(5 * time.Second)
		} else {
			break
		}

		if retryCount == maxRetries-1 {
			log.Println("Max retries reached. Could not reconnect to the relay.")
		}
	}

	fmt.Println("done")
}
