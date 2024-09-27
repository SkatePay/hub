package api

import (
	"encoding/json"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type TokenClaims struct {
	AllowedBucket string `json:"bucket"`
	jwt.StandardClaims
}

var jwtKey = []byte("your-secret-key")

func GenerateToken(bucket string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &TokenClaims{
		AllowedBucket: bucket,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func handleTokenRequest(w http.ResponseWriter, r *http.Request) {
	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		http.Error(w, "Bucket required", http.StatusBadRequest)
		return
	}

	tokenString, err := GenerateToken(bucket)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
