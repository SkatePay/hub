package api

import (
	"encoding/json"
	"net/http"
	"os"
)

func handleKeysRequest(w http.ResponseWriter, r *http.Request) {
	s3AccessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	s3SecretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")

	json.NewEncoder(w).Encode(map[string]string{
		"S3_ACCESS_KEY_ID":     s3AccessKeyID,
		"S3_SECRET_ACCESS_KEY": s3SecretAccessKey,
	})
}
