package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
		{ID: uuid.New(), Name: "Kickflips", Icon: "🏆", Coordinate: Coordinate{Latitude: 33.987210164306546, Longitude: -118.47545038626512},
			ChannelId: "6e1fd586debdd7b0b8a13b56dc6c4087ffbf4299f0c2a5845df39f231e5aa276"},
		{ID: uuid.New(), Name: "Cleaning Crew", Icon: "🧹", Coordinate: Coordinate{Latitude: 33.98703813963489, Longitude: -118.47560084057935},
			ChannelId: "bad52e45b13ebc9603baaf56b5b9adb5599a634a4a32037ec13427dcd88e7657"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

func Start() {
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/leads", leadsHandles)

	// curl -X GET "https://localhost:3001/token?bucket=skateconnect" -H "Content-Type: application/json"
	http.HandleFunc("/token", handleTokenRequest)

	fmt.Printf("Starting server on port 3001\n")
	if err := http.ListenAndServe(":3001", nil); err != nil {
		log.Fatal(err)
	}
}
