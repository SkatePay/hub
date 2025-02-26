package api

import "github.com/google/uuid"

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
