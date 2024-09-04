package main

import (
	"flag"
	"hub/publisher"
	"hub/subscriber"
	"hub/workers"
	"log"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/joho/godotenv"
)

const USAGE = `hub

Usage:
  hub listen
  hub spawn
  hub publish

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
	case opts["listen"].(bool):
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		nsec := os.Getenv("HUB_NSEC")
		npub := os.Getenv("HUB_NPUB")

		subscriber.Subscribe(nsec, npub)
	case opts["spawn"].(bool):
		workers.Create_Worker()
	case opts["publish"].(bool):
		// publisher
		npub_Receiver := "npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r" // 🌊 primal
		// npub_Receiver = ""  // 🛹 skatepark

		// publisher.Publish()
		publisher.Publish_Encrypted(npub_Receiver)
	}
}
