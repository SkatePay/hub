package subscriber

import (
	"encoding/json"
	"hub/nostr/weather"
	"strings"

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

func ChatBot(channelId string, nsec string, npub string) {
	url := os.Getenv("HUB_RELAY")

	_, pk, _ := nip19.Decode(npub)
	_, sk, _ := nip19.Decode(nsec)

	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, url)
	if err != nil {
		panic(err)
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

	// ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		panic(err)
	}

	var events []*nostr.Event
	var processingStoredEvents bool

	for {
		select {
		case event, ok := <-sub.Events:
			if !ok {
				return
			}
			if !processingStoredEvents {
				events = append(events, event)
			} else {
				// Process new events as they come in
				provider := &DefaultProvider{
					Relay:      relay,
					ChannelId:  channelId,
					PrivateKey: sk.(string),
					PublicKey:  pk.(string),
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
			fmt.Println("done")
			return
		}
	}
}

// ContentStructure represents the structure of the content
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
