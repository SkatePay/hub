package main

import (
	"flag"
	"fmt"
	"hub/api"
	groupbot "hub/bots/group-bot"
	"hub/nostr/workers"
	"hub/solana"
	"log"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/joho/godotenv"
)

const USAGE = `hub

Usage:
  hub api
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

	nsec, npub, channelID, relayURL := getEnvVariables()

	// Command Execution
	switch {
	case opts["api"].(bool):
		startAPI()

	case opts["group-bot"].(bool):
		startGroupBot(nsec, npub, channelID, relayURL)

	case opts["quick_identity"].(bool):
		workers.Quick_Identity()

	case opts["broadcast"].(bool):
		workers.Broadcast()

	case opts["scan"].(bool):
		workers.Scan()

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

func getEnvVariables() (string, string, string, string) {
	nsec := os.Getenv("HUB_NSEC")
	npub := os.Getenv("HUB_NPUB")
	relayURL := os.Getenv("HUB_RELAY")
	channelID := os.Getenv("HUB_CHANNEL_ID")

	if nsec == "" || npub == "" || channelID == "" || relayURL == "" {
		log.Fatal("❌ Missing required environment variables: HUB_NSEC, HUB_NPUB, HUB_CHANNEL_ID, HUB_RELAY")
	}

	return nsec, npub, channelID, relayURL
}

// ✅ Command Execution Functions

func startAPI() {
	log.Println("🚀 Starting API service...")
	api.Start()
}

func startGroupBot(nsec, npub, channelID string, relayURL string) {
	log.Println("🤖 Starting Group Chat Bot...")

	bot, err := groupbot.NewGroupBot(nsec, npub, relayURL, channelID)
	if err != nil {
		log.Fatalf("❌ Failed to initialize DMBot: %v", err)
	}

	bot.Start()
}
