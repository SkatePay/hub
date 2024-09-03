package subscriber

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func Subscribe(nsec string, npub string) {
	fmt.Println(nsec)
	fmt.Println(npub)
	fmt.Println()

	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, "wss://relay.primal.net")
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening for nostr events...")

	var filters nostr.Filters
	_, v1, _ := nip19.Decode(npub)
	_, v2, _ := nip19.Decode("npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r") // 🌊 primal

	filters = []nostr.Filter{{
		Kinds:   []int{nostr.KindEncryptedDirectMessage},
		Authors: []string{v1.(string), v2.(string)},
	}}

	// ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		panic(err)
	}

	for ev := range sub.Events {
		// handle returned event.
		// channel will stay open until the ctx is cancelled (in this case, context timeout)
		fmt.Println(ev.ID)
		fmt.Println(ev.Content)
	}

	fmt.Println("done")
}
