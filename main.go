package main

import (
	"flag"
	"hub/nostr/publisher"
	"hub/nostr/subscriber"
	"hub/nostr/workers"
	s_p "hub/solana/publisher"
	"log"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/joho/godotenv"
)

const USAGE = `hub

Usage:
  hub nostr_listen
  hub nostr_spawn
  hub nostr_publish
  hub solana_publish

Specify <content> as '-' to make the publish or message command read it
from stdin.
`

func main() {
	log.SetPrefix("<> ")

	opts, err := docopt.ParseArgs(USAGE, flag.Args(), "")

	if err != nil {
		return
	}

	switch {
	case opts["nostr_listen"].(bool):
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		nsec := os.Getenv("HUB_NSEC")
		npub := os.Getenv("HUB_NPUB")

		subscriber.Subscribe(nsec, npub)

	case opts["nostr_spawn"].(bool):
		workers.Create_Worker()

	case opts["nostr_publish"].(bool):
		// publisher
		npub_Receiver := "npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r" // 🌊 primal
		npub_Receiver = "npub1amffkjlqudax8egy8e587ajdh4xekj0y0vktj2te4mx8cnnekfxs8yx299"  // 🛹 skatepark

		// publisher.Publish()
		publisher.Publish_Encrypted(npub_Receiver)

	case opts["solana_publish"].(bool):
		s_p.Publish()
	}
}
