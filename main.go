package main

import (
	"flag"
	"hub/api"
	"hub/nostr/publisher"
	hub "hub/nostr/subscriber"
	"hub/nostr/workers"
	s_p "hub/solana/publisher"
	"log"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/joho/godotenv"
)

const USAGE = `hub

Usage:
  hub api
  hub support
  hub chatbot
  hub quick_identity
  hub broadcast
  hub scan
  hub ping
  hub quick_wallet

Specify <content> as '-' to make the publish or message command read it
from stdin.
`

func main() {
	log.SetPrefix("<> ")

	opts, err := docopt.ParseArgs(USAGE, flag.Args(), "")

	if err != nil {
		return
	}

	err = godotenv.Load()
	if err != nil {
		// log.Fatal("Error loading .env file")
	}

	nsec := os.Getenv("HUB_NSEC")
	npub := os.Getenv("HUB_NPUB")
	channelId := os.Getenv("HUB_CHANNEL_ID")

	switch {
	// Start API
	case opts["api"].(bool):
		api.Start()

	// Start Support
	case opts["support"].(bool):
		hub.TechSupport(nsec, npub, channelId)

	// Monitor Public Chat
	case opts["chatbot"].(bool):
		hub.ChatBot(nsec, npub, channelId)

	// Nostr Utilities
	case opts["quick_identity"].(bool):
		workers.Quick_Identity()

	case opts["broadcast"].(bool):
		workers.Broadcast()

	case opts["scan"].(bool):
		workers.Scan()

	case opts["ping"].(bool):
		// publisher
		npub_Receiver := "npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r" // 🌊 primal

		// publisher.Publish()
		publisher.Publish_Encrypted(npub_Receiver, "🙃")

	// Solana Utilities
	case opts["quick_wallet"].(bool):
		s_p.Quick_Wallet()
	}
}
