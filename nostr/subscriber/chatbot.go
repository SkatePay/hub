package subscriber

import (
	"encoding/json"
	"hub/nostr/weather"
	"strings"
	"time"

	"context"
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Message struct {
	Content string `json:"content"`
	Kind    string `json:"kind"`
}

type DefaultProvider struct {
	Relay      *nostr.Relay
	ChannelId  string
	PublicKey  string
	PrivateKey string
}

func (p *DefaultProvider) GetRelay() *nostr.Relay {
	return p.Relay
}

func (p *DefaultProvider) GetChannelId() string {
	return p.ChannelId
}

func (p *DefaultProvider) GetPrivateKey() string {
	return p.PrivateKey
}

func (p *DefaultProvider) GetPublicKey() string {
	return p.PublicKey
}

type RelayProvider interface {
	GetRelay() *nostr.Relay
	GetChannelId() string
	GetPrivateKey() string
	GetPublicKey() string
}

func ProcessEvent(provider RelayProvider, ctx context.Context, event nostr.Event) {
	relay := provider.GetRelay()
	channelId := provider.GetChannelId()
	pk := provider.GetPublicKey()
	sk := provider.GetPrivateKey()

	npub, _ := nip19.EncodePublicKey(event.PubKey)
	suffix := npub[len(npub)-3:]
	username := fmt.Sprintf("skate-%s", suffix)

	var message Message
	err := json.Unmarshal([]byte(event.Content), &message)
	if err != nil {
		fmt.Println(username, event.Content)
	} else {
		fmt.Println(username, message.Content)

		if strings.Contains(message.Content, "!weather") {
			content := weather.GetReport()

			var ev nostr.Event
			var tags nostr.Tags

			tags = append(tags, nostr.Tag{"e", channelId, relay.URL, "root"})

			ev = nostr.Event{
				PubKey:    pk,
				CreatedAt: nostr.Now(),
				Kind:      nostr.KindChannelMessage,
				Content:   content,
				Tags:      tags,
			}

			ev.Sign(sk)

			if err := relay.Publish(ctx, ev); err != nil {
				fmt.Println(err)
			}
		}
	}
}

// Max number of retries to reconnect
const maxRetries = 5

// ReconnectDelay is the delay between reconnection attempts
const reconnectDelay = 5 * time.Second

func ChatBot(nsec string, npub string, channelId string) {
	url := os.Getenv("HUB_RELAY")

	_, pk, _ := nip19.Decode(npub)
	_, sk, _ := nip19.Decode(nsec)

	ctx := context.Background()

	// Function to establish the connection and subscribe to the relay
	connectAndSubscribe := func() (*nostr.Subscription, *nostr.Relay, error) {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
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

		// Once connected, process events
		processEvents(ctx, sub, relay, pk.(string), sk.(string), channelId)

		// If processing completes without errors, break the retry loop
		break
	}

	fmt.Println("Max retries reached. Could not reconnect to the relay.")
}

// Function to process events and handle relay disconnections
func processEvents(ctx context.Context, sub *nostr.Subscription, relay *nostr.Relay, pk, sk, channelId string) {
	var events []*nostr.Event
	var processingStoredEvents bool

	for {
		select {
		case event, ok := <-sub.Events:
			if !ok {
				fmt.Println("Relay disconnected, trying to reconnect...")
				return // Exit to trigger reconnection logic
			}
			if !processingStoredEvents {
				events = append(events, event)
			} else {
				// Process new events as they come in
				provider := &DefaultProvider{
					Relay:      relay,
					ChannelId:  channelId,
					PrivateKey: sk,
					PublicKey:  pk,
				}

				ProcessEvent(provider, ctx, *event)
			}

		case <-sub.EndOfStoredEvents:
			if !processingStoredEvents {
				fmt.Println("End of stored events received")
				processingStoredEvents = true
				// Process stored events in reverse order
				for i := len(events) - 1; i >= 0; i-- {
					event := events[i]

					npub, _ := nip19.EncodePublicKey(event.PubKey)
					suffix := npub[len(npub)-3:]
					username := fmt.Sprintf("skate-%s", suffix)

					var message Message
					err := json.Unmarshal([]byte(event.Content), &message)
					if err != nil {
						fmt.Println(username, event.Content)
					} else {
						fmt.Println(username, message.Content)
					}
				}
				events = nil // Reset the buffer for new events
			}

		case <-relay.Context().Done():
			fmt.Println("Relay context done, closing connection...")
			return // Exit to trigger reconnection logic
		}
	}
}

type ContentStructure struct {
	Content string `json:"content"`
	Kind    string `json:"kind"`
}

func createMessage(npub string) string {
	// Create an instance of ContentStructure
	message := ContentStructure{
		Content: npub,
		Kind:    "subscriber",
	}

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return ""
	}

	// Convert the JSON bytes into a string
	return string(jsonData)
}

func Announce(channelId string, npubForSubscriber string, nsecForPublisher string, npubForPublisher string) {
	nrelay := os.Getenv("HUB_RELAY")
	eventId := os.Getenv("HUB_CHANNEL_ID")

	_, pk, _ := nip19.Decode(npubForPublisher)
	_, sk, _ := nip19.Decode(nsecForPublisher)

	var ev nostr.Event
	var tags nostr.Tags

	tags = append(tags, nostr.Tag{"e", eventId, nrelay, "root"})

	ev = nostr.Event{
		PubKey:    pk.(string),
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindChannelMessage,
		Content:   createMessage(npubForSubscriber),
		Tags:      tags,
	}

	ev.Sign(sk.(string))

	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, nrelay)
	if err != nil {
		panic(err)
	}
	if err := relay.Publish(ctx, ev); err != nil {
		fmt.Println(err)
	}
}
