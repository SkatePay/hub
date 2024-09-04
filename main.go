package main

import (
	"hub/subscriber"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// // workers

	// workers.Create_Worker()

	// return

	// subscriber
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	nsec := os.Getenv("HUB_NSEC")
	npub := os.Getenv("HUB_NPUB")

	subscriber.Subscribe(nsec, npub)

	// // publisher
	// npub_Receiver := "npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r" // 🌊 primal
	// // npub_Receiver = ""  // 🛹 skatepark

	// // publisher.Publish()
	// publisher.Publish_Encrypted(npub_Receiver)
}
