package main

import (
	"flag"
	"fmt"
	"hub/api"
	groupbot "hub/bots/group-bot"
	"hub/core"
	"hub/nostr/workers"
	"hub/solana"
	"hub/utils"
	"log"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/joho/godotenv"
)

const USAGE = `hub

Usage:
  hub api
  hub dm-bot
  hub group-bot
  hub quick_identity
  hub broadcast
  hub scan
  hub ping
  hub quick_wallet

Specify <content> as '-' to make the publish or message command read it
from stdin.
`

func main() {
	initializeLogging()
	loadEnvVariables()

	opts, err := docopt.ParseArgs(USAGE, flag.Args(), "")
	if err != nil {
		log.Fatalf("❌ Failed to parse CLI arguments: %v", err)
	}

	nsec, npub, channelID := getEnvVariables()

	// Command Execution
	switch {
	case opts["api"].(bool):
		startAPI()

	case opts["dm-bot"].(bool):
		startDMBot(nsec, npub, channelID)

	case opts["group-bot"].(bool):
		startGroupBot(nsec, npub, channelID)

	case opts["quick_identity"].(bool):
		workers.Quick_Identity()

	case opts["broadcast"].(bool):
		workers.Broadcast()

	case opts["scan"].(bool):
		workers.Scan()

	case opts["ping"].(bool):
		pingExample()

	case opts["quick_wallet"].(bool):
		solana.Quick_Wallet()

	default:
		fmt.Println("❗ Invalid command. Use '--help' for usage instructions.")
	}
}

// ✅ Utility Functions

func initializeLogging() {
	log.SetPrefix("[hub] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func loadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Could not load .env file. Using system environment variables...")
	}
}

func getEnvVariables() (string, string, string) {
	nsec := os.Getenv("HUB_NSEC")
	npub := os.Getenv("HUB_NPUB")
	channelID := os.Getenv("HUB_CHANNEL_ID")

	if nsec == "" || npub == "" || channelID == "" {
		log.Fatal("❌ Missing required environment variables: HUB_NSEC, HUB_NPUB, or HUB_CHANNEL_ID")
	}

	return nsec, npub, channelID
}

// ✅ Command Execution Functions

func startAPI() {
	log.Println("🚀 Starting API service...")
	api.Start()
}

func startDMBot(nsec, npub, channelID string) {
	log.Println("🤖 Starting Direct Message Bot...")
	core.TechSupport(nsec, npub, channelID)
}

func startGroupBot(nsec, npub, channelID string) {
	log.Println("🤖 Starting Group Chat Bot...")
	groupbot.GroupBot(nsec, npub, channelID)
}

func pingExample() {
	npubReceiver := "npub1uxp7mwl2mtetc4qmr0y6ck0p0y50c3zhglzzwvvdzf6dvpsjtvvq9gs05r" // 🌊 primal
	log.Println("📡 Sending encrypted ping message...")
	utils.PublishEncrypted(npubReceiver, "🙃")
}
