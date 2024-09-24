package subscriber

import (
	"context"
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func PublicChat(nsec string, npub string) {
	url := os.Getenv("HUB_RELAY")
	eventId := os.Getenv("HUB_CHANNEL_ID")

	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, url)
	if err != nil {
		panic(err)
	}

	fmt.Println("📡", eventId, " connected")
	fmt.Println("🇺🇸", npub, "online")
	fmt.Println()

	tags := make(map[string][]string)
	tags["e"] = []string{eventId}

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
				npub, _ := nip19.EncodePublicKey(event.PubKey)
				suffix := npub[len(npub)-3:]
				username := fmt.Sprintf("skate-%s", suffix)
				fmt.Println(username, event.Content)
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
					fmt.Println(username, event.Content)
				}
				events = nil // Reset the buffer for new events
			}

		case <-relay.Context().Done():
			fmt.Println("done")
			return
		}
	}
}
