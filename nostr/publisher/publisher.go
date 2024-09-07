package publisher

import (
	"context"
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func Publish() {
	sk := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(sk)

	ev := nostr.Event{
		PubKey:    pub,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      nil,
		Content:   "Hello, World!",
	}

	ev.Sign(sk)

	ctx := context.Background()
	for _, url := range []string{"wss://relay.primal.net"} {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if err := relay.Publish(ctx, ev); err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("published to %s\n", url)
	}
}

func Publish_Encrypted(npub_Receiver string, message string) {
	fmt.Println("Publishing encrypted message to", npub_Receiver)

	var sk string
	nsec := os.Getenv("HUB_NSEC")
	npub := os.Getenv("HUB_NPUB")

	if nsec == "" || npub == "" {
		fmt.Println("Generating new keys since HUB_NSEC or HUB_NPUB is not set:")

		sk = nostr.GeneratePrivateKey()

		pk, _ := nostr.GetPublicKey(sk)

		nsec, _ = nip19.EncodePrivateKey(sk)
		npub, _ = nip19.EncodePublicKey(pk)

		fmt.Println("Generated nsec:", nsec)
		fmt.Println("Generated npub:", npub)
		fmt.Println()
	} else {
		fmt.Println("Using existing keys from environment:")
		fmt.Println("npub:", npub)
		fmt.Println()
	}

	if _, v, err := nip19.Decode(npub_Receiver); err == nil {
		receiverKey := v.(string)

		_, sk, err := nip19.Decode(nsec)
		if err != nil {
			fmt.Println("Error decoding private key:", err)
			return
		}

		shared, _ := nip04.ComputeSharedSecret(receiverKey, sk.(string))

		encryptedMessage, _ := nip04.Encrypt(message, shared)

		var tags nostr.Tags
		tags = append(tags, nostr.Tag{"p", receiverKey})

		ev := nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      nostr.KindEncryptedDirectMessage,
			Tags:      tags,
			Content:   encryptedMessage,
		}
		ev.Sign(sk.(string))

		ctx := context.Background()
		for _, url := range []string{"wss://relay.primal.net"} {
			relay, err := nostr.RelayConnect(ctx, url)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if err := relay.Publish(ctx, ev); err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("published to %s\n", url)
		}
	} else {
		panic(err)
	}
}
