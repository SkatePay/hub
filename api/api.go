package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Status struct {
	Status string `json:"status"`
}

type Lead struct {
	Name      string  `json:"name"`
	Icon      string  `json:"icon"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ChannelId string  `json:"channelId"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := Status{Status: "online"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func leadsHandles(w http.ResponseWriter, r *http.Request) {
	leads := []Lead{
		{Name: "Kickflip Contest", Icon: "🏆", Latitude: 33.987210164306546, Longitude: -118.47545038626512, ChannelId: "6e1fd586debdd7b0b8a13b56dc6c4087ffbf4299f0c2a5845df39f231e5aa276"},
		// Add more Lead entries as needed
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leads)
}

func Start() {
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/leads", leadsHandles)

	fmt.Printf("Starting server on port 3001\n")
	if err := http.ListenAndServe(":3001", nil); err != nil {
		log.Fatal(err)
	}
}
