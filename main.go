package main

import (
	"flag"
	"fmt"
	"hub/api"
	"hub/nostr/workers"
	"hub/solana"
	"log"

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

	// Command Execution
	switch {
	case opts["api"].(bool):
		startAPI()

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

// ✅ Command Execution Functions

func startAPI() {
	log.Println("🚀 Starting API service...")
	api.Start()
}
