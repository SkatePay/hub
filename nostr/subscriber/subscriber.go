package subscriber

import (
	"context"
	"fmt"
	"hub/nostr/publisher"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
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

	_, sk, _ := nip19.Decode(nsec)

	for ev := range sub.Events {
		shared, _ := nip04.ComputeSharedSecret(ev.PubKey, sk.(string))

		npub, _ := nip19.EncodePublicKey(ev.PubKey)
		fmt.Println(npub)

		ciphertext := ev.Content
		plaintext, _ := nip04.Decrypt(ciphertext, shared)

		fmt.Println(plaintext)

		fmt.Println()

		if plaintext == "🙂" {
			publisher.Publish_Encrypted(npub, "🙃")
		}
	}

	fmt.Println("done")
}
