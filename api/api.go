package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Message represents a simple message structure
type Message struct {
	Status string `json:"status"`
}

// Handler function to respond with JSON
func messageHandler(w http.ResponseWriter, r *http.Request) {
	message := Message{Status: "online"}

	// Set the header to indicate JSON response
	w.Header().Set("Content-Type", "application/json")

	// Encode the message to JSON and write it to the response
	json.NewEncoder(w).Encode(message)
}

func Start() {
	// Define the route and handler
	http.HandleFunc("/status", messageHandler)

	// Start the server on port 3001
	fmt.Printf("Starting server on port 3001\n")
	if err := http.ListenAndServe(":3001", nil); err != nil {
		log.Fatal(err)
	}
}
