package workers

import (
	"context"
	"fmt"
	"hub/nostr/weather"
	"os"

	"github.com/joho/godotenv"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func Broadcast() {
	err := godotenv.Load()
	if err != nil {
		// log.Fatal("Error loading .env file")
	}

	nsec := os.Getenv("HUB_NSEC")
	npub := os.Getenv("HUB_NPUB")
	nrelay := os.Getenv("HUB_RELAY")

	_, pk, _ := nip19.Decode(npub)
	_, sk, _ := nip19.Decode(nsec)

	// Start relay listener
	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, nrelay)
	if err != nil {
		panic(err)
	}

	// // Create a new channel
	var content string
	var ev nostr.Event
	var tags nostr.Tags

	createChannel := func() {
		content = "{\"name\": \"Demo Channel\", \"about\": \"A test channel.\", \"picture\": \"https://placekitten.com/200/200\", \"relays\": [\"wss://relay.primal.net\"]}"

		ev = nostr.Event{
			PubKey:    pk.(string),
			CreatedAt: nostr.Now(),
			Kind:      nostr.KindChannelCreation,
			Content:   content,
		}

		ev.Sign(sk.(string))

		if err := relay.Publish(ctx, ev); err != nil {
			fmt.Println(err)
		}
	}

	updateChannel := func() {
		content = "{\"name\": \"Updated Demo Channel\", \"about\": \"Updating a test channel.\", \"picture\": \"https://placekitten.com/201/201\", \"relays\": [\"wss://relay.primal.net\"]}"

		tags = append(tags, nostr.Tag{"e", "dde50a64b7aab5cc36c9e2944b452ecbec910dc52ba1a9078028dc227564f01f", nrelay})

		fmt.Printf("tags: %v\n", tags)

		ev = nostr.Event{
			PubKey:    pk.(string),
			CreatedAt: nostr.Now(),
			Kind:      nostr.KindChannelMetadata,
			Content:   content,
			Tags:      tags,
		}

		ev.Sign(sk.(string))

		if err := relay.Publish(ctx, ev); err != nil {
			fmt.Println(err)
		}
	}

	createMessage := func() {
		// Root
		// {
		// 	"content": <string>,
		// 	"tags": [["e", <kind_40_event_id>, <relay-url>, "root"]],
		// 	// other fields...
		// }

		// Reply
		// {
		// 	"content": <string>,
		// 	"tags": [
		// 		["e", <kind_40_event_id>, <relay-url>, "root"],
		// 		["e", <kind_42_event_id>, <relay-url>, "reply"],
		// 		["p", <pubkey>, <relay-url>],
		// 		// rest of tags...
		// 	],
		// 	// other fields...
		// }

		countryCode := "US"
		zipCode := "90291"

		values := weather.GetWeather(countryCode, zipCode)

		if len(values) == 0 {
			fmt.Println("No weather data found")
			return
		}
		chunks := []interface{}{"Current Weather:", values[0], "°F"}

		content := fmt.Sprintf("%v %v %v ☀️", chunks...)

		eventId := os.Getenv("HUB_CHANNEL_ID")

		tags = append(tags, nostr.Tag{"e", eventId, nrelay, "root"})

		ev = nostr.Event{
			PubKey:    pk.(string),
			CreatedAt: nostr.Now(),
			Kind:      nostr.KindChannelMessage,
			Content:   content,
			Tags:      tags,
		}

		ev.Sign(sk.(string))

		if err := relay.Publish(ctx, ev); err != nil {
			fmt.Println(err)
		}

		fmt.Printf("✉️ ➡️ %s\n", relay)
	}

	if false {
		createChannel()
		updateChannel()
	}

	createMessage()
}

func Scan() {
	err := godotenv.Load()
	if err != nil {
		// log.Fatal("Error loading .env file")
	}

	npub := os.Getenv("HUB_NPUB")
	nrelay := os.Getenv("HUB_RELAY")

	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, nrelay)
	if err != nil {
		panic(err)
	}

	var filters nostr.Filters
	if _, v, err := nip19.Decode(npub); err == nil {
		pub := v.(string)
		fmt.Printf("pub: %s\n", pub)
		pub = "b41bd8a39b6d5889c4759f0f35716b350cc170bf6d1d2d4c23937ddb6929af65"
		id := "22e4dd5d387b8cf8f96fa159af04edf845d6526149e0e93b8a0f6b23c97af4a2"
		filters = []nostr.Filter{{
			IDs:   []string{id},
			Kinds: []int{nostr.KindChannelCreation, nostr.KindChannelMetadata},
			// Authors: []string{pub},
		}}
	} else {
		panic(err)
	}

	fmt.Println("Listening for nostr events...")

	// ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	// defer cancel()

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		panic(err)
	}

	for ev := range sub.Events {
		fmt.Println(ev.ID, ev)
		fmt.Println()
	}
}
