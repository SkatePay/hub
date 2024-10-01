package subscriber

import (
	"context"
	"fmt"
	"hub/nostr/publisher"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func getUsername(input string) string {
	input = strings.TrimSuffix(input, ".")
	length := len(input)
	return input[length-10:]
}

func Subscribe(nsecForHost string, npubForHost string, channelId string) {
	fmt.Println(npubForHost, "online")
	fmt.Println("channelId", channelId)
	fmt.Println()

	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, "wss://relay.primal.net")
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening for nostr events...")

	_, v1, _ := nip19.Decode(npubForHost)

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

	_, sk, _ := nip19.Decode(nsecForHost)

	for ev := range sub.Events {
		shared, _ := nip04.ComputeSharedSecret(ev.PubKey, sk.(string))

		npub, _ := nip19.EncodePublicKey(ev.PubKey)
		fmt.Println()

		ciphertext := ev.Content
		plaintext, _ := nip04.Decrypt(ciphertext, shared)

		fmt.Println(npub, ":", plaintext)
		if plaintext == "🙂" {
			publisher.Publish_Encrypted(npub, "🙃")
		}

		if strings.Contains(plaintext, "Hi, I would like to report ") {
			message := fmt.Sprintf("Could you elaborate on the problem you're encountering with %s? Additional details would greatly assist in resolving your issue. In the meanwhile, feel free to mute the user if that's necessary.", getUsername(plaintext))
			publisher.Publish_Encrypted(npub, message)
		}

		if strings.Contains(plaintext, "I'm online.") {
			message := "Welcome to SkateConnect, skater! If you have any questions or need to report a bug do not hesitate to get in touch with us."
			publisher.Publish_Encrypted(npub, message)

			// Tell everyone who just joined Public Chat
			Announce(channelId, npub, nsecForHost, npubForHost)
		}
	}

	fmt.Println("done")
}
