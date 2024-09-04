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

	_, v1, _ := nip19.Decode(npub)
	// _, v2, _ := nip19.Decode("npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r") // 🌊 primal
	// _, v3, _ := nip19.Decode("npub1vzjyahj8zach3ydfv2fmqk3adgwvctpcnr69vc98uza9cw955tas3ntdzv") // 🛹 skatechat

	tags := make(map[string][]string)
	tags["p"] = []string{v1.(string)}

	filters := []nostr.Filter{{
		Kinds: []int{nostr.KindEncryptedDirectMessage},
		// Authors: []string{v1.(string), v2.(string)},
		Tags:  tags,
		Limit: 1,
	}}

	// ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		panic(err)
	}

	for ev := range sub.Events {
		// channel will stay open until the ctx is cancelled (in this case, context timeout)
		fmt.Println(ev.Content)
		fmt.Println(ev.PubKey)
		fmt.Println(ev.Tags)
		fmt.Println()
	}

	fmt.Println("done")
}
