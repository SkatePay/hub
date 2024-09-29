package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func handleKeysRequest(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	S3_ACCESS_KEY_ID := os.Getenv("S3_ACCESS_KEY_ID")
	S3_SECRET_ACCESS_KEY := os.Getenv("S3_SECRET_ACCESS_KEY")

	json.NewEncoder(w).Encode(map[string]string{
		"S3_ACCESS_KEY_ID":     S3_ACCESS_KEY_ID,
		"S3_SECRET_ACCESS_KEY": S3_SECRET_ACCESS_KEY,
	})
}
