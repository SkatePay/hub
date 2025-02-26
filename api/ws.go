package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Message struct for JSON encoding
type Message struct {
	Text string `json:"text"`
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (for dev, restrict in production)
	},
}

// WebSocket clients map
var clients = make(map[*websocket.Conn]bool)

// Broadcast channel to send messages
var broadcast = make(chan Message)

// InitWebSocket starts the message broadcasting loop in a separate goroutine
func InitWebSocket() {
	go handleMessages()
}

// HandleWebSocket handles WebSocket connections.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	log.Println("New WebSocket client connected")

	// Listen for messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket error:", err)
			delete(clients, conn)
			break
		}

		log.Printf("Received: %s\n", msg)

		// Ensure valid JSON before broadcasting
		var receivedMsg Message
		if err := json.Unmarshal(msg, &receivedMsg); err != nil {
			log.Println("Invalid JSON received:", err)
			continue
		}

		// Send structured message to the broadcast channel
		broadcast <- receivedMsg
	}
}

// handleMessages listens for messages on the broadcast channel and sends them to all clients
func handleMessages() {
	for {
		msg := <-broadcast
		jsonMsg, err := json.Marshal(msg) // Ensure outgoing message is JSON
		if err != nil {
			log.Println("Error marshalling JSON:", err)
			continue
		}

		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
				log.Println("Error broadcasting:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
