package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/rs/cors"
)

// Status represents the API status response.
type Status struct {
	Status string `json:"status"`
}

// Coordinate represents a geographic coordinate.
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Lead represents a lead with its details.
type Lead struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Icon       string     `json:"icon"`
	Coordinate Coordinate `json:"coordinate"`
	ChannelID  string     `json:"channelId"`
}

// statusHandler handles requests to the /status endpoint.
func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := Status{Status: "online"}
	sendJSONResponse(w, status)
}

// leadsHandler handles requests to the /leads endpoint.
func leadsHandler(w http.ResponseWriter, r *http.Request) {
	leads := []Lead{
		{
			ID:   uuid.New(),
			Name: "Public Chat",
			Icon: "📡",
			Coordinate: Coordinate{
				Latitude:  33.98686098062241,
				Longitude: -118.4754199190118,
			},
			ChannelID: "92ef3ac79a8772ddf16a2e74e239a67bc95caebdb5bd59191c95cf91685dfc8e",
		},
	}
	sendJSONResponse(w, leads)
}

// sendJSONResponse sends a JSON response with the correct headers.
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Start initializes and starts the HTTP server.
func Start() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/leads", leadsHandler)
	mux.HandleFunc("/keys", handleKeysRequest)
	mux.HandleFunc("/channel/", handleChannelVideos)
	mux.HandleFunc("/token", handleTokenRequest)

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://skatepark.chat", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            false,
	})

	// Wrap the ServeMux with the CORS middleware
	handler := c.Handler(mux)

	// Get the port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server with CORS middleware
	fmt.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
