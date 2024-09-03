package main

import (
	"hub/subscriber"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func main() {
	sk := nostr.GeneratePrivateKey()
	// pk, _ := nostr.GetPublicKey(sk)

	nsec, _ := nip19.EncodePrivateKey(sk)
	// npub, _ := nip19.EncodePublicKey(pk)

	npub := "npub15770rt0a8knm22mka87zw26hjezvfmn4hvl2rkn7zq7pfm2md4mqjxfkc9"

	subscriber.Subscribe(nsec, npub)

	// npub_Receiver := "npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r" // 🌊 primal
	// // npub_Receiver = ""  // 🛹 skatepark

	// // publisher.Publish()
	// publisher.Publish_Encrypted(npub_Receiver)
}
