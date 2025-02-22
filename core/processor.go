package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

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
