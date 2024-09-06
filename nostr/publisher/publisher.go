package publisher

import (
	"context"
	"fmt"

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

func Publish_Encrypted(npub_Receiver string) {
	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)

	nsec, _ := nip19.EncodePrivateKey(sk)
	npub, _ := nip19.EncodePublicKey(pk)

	fmt.Println(nsec)
	fmt.Println(npub)
	fmt.Println()

	if _, v, err := nip19.Decode(npub_Receiver); err == nil {
		receiverKey := v.(string)

		shared, _ := nip04.ComputeSharedSecret(receiverKey, sk)

		message := "🔋"
		encryptedMessage, _ := nip04.Encrypt(message, shared)

		var tags nostr.Tags
		tags = append(tags, nostr.Tag{"p", receiverKey})

		ev := nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      nostr.KindEncryptedDirectMessage,
			Tags:      tags,
			Content:   encryptedMessage,
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
	} else {
		panic(err)
	}
}
