package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type Status struct {
	Status string `json:"status"`
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Lead struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Icon       string     `json:"icon"`
	Coordinate Coordinate `json:"coordinate"`
	ChannelId  string     `json:"channelId"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := Status{Status: "online"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func leadsHandles(w http.ResponseWriter, r *http.Request) {
	leads := []Lead{
		{ID: uuid.New(), Name: "Public Chat", Icon: "📡", Coordinate: Coordinate{Latitude: 33.98685829544777, Longitude: -118.47541361363558},
			ChannelId: "ab3561547df90fc8840022577ef039f61383daae9adb1960d1485968a5ca39fa"},
		{ID: uuid.New(), Name: "Kickflips", Icon: "🏆", Coordinate: Coordinate{Latitude: 33.98722350529692, Longitude: -118.47543510467679},
			ChannelId: "006e90a428579dfe5cd894ad990ac9d7039f7c50a2d9c5184199cdfba4e5b635"},
		{ID: uuid.New(), Name: "Cleaning Crew", Icon: "🧹", Coordinate: Coordinate{Latitude: 33.98703813963489, Longitude: -118.47560084057935},
			ChannelId: "bad52e45b13ebc9603baaf56b5b9adb5599a634a4a32037ec13427dcd88e7657"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

// CORS Middleware to handle CORS requests
func enableCORS(w http.ResponseWriter, _ *http.Request) {
	// Check the environment (COPILOT_ENVIRONMENT_NAME)
	env := os.Getenv("COPILOT_ENVIRONMENT_NAME")
	var allowedOrigin string

	// Set allowed origin based on the environment
	if env == "production" {
		// In production, allow only requests from the production domain
		allowedOrigin = "https://skatepark.chat"
	} else {
		// In non-production (e.g., dev or local), allow localhost for front-end dev
		allowedOrigin = "http://localhost:3000"
	}

	// Set CORS headers dynamically based on the environment
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

	// Allow specific HTTP methods (GET, POST, OPTIONS)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	// Allow headers that are sent in the request
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Allow credentials (e.g., cookies, authorization headers)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// Middleware function to apply CORS headers
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)

		// Handle preflight requests (OPTIONS method)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler if it's not a preflight request
		next.ServeHTTP(w, r)
	})
}

func Start() {
	// curl -X GET "https://localhost:3001/status" -H "Content-Type: application/json"

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/leads", leadsHandles)
	http.HandleFunc("/keys", handleKeysRequest)

	// curl "http://localhost:3001/channel/ab3561547df90fc8840022577ef039f61383daae9adb1960d1485968a5ca39fa" -H "Content-Type: application/json"
	http.HandleFunc("/channel/", handleChannelVideos)

	// curl -X GET "https://localhost:3001/token?bucket=skateconnect" -H "Content-Type: application/json"
	http.HandleFunc("/token", handleTokenRequest)

	// Get the port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(":"+port, withCORS(http.DefaultServeMux)); err != nil {
		log.Fatal(err)
	}
}
