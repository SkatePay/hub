package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

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

func getS3VideosByChannel(channelId string) ([]string, error) {
	fmt.Print(channelId)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	svc := s3.New(sess)
	bucket := os.Getenv("S3_BUCKET")

	// Use S3 ListObjectsV2 API to list all objects with tag channelId=<channelId>
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	result, err := svc.ListObjectsV2(input)
	if err != nil {
		return nil, fmt.Errorf("unable to list items in bucket %q, %v", bucket, err)
	}

	var videos []string
	for _, item := range result.Contents {
		// Check if the item is tagged with the correct channelId
		tagsInput := &s3.GetObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    item.Key,
		}

		tagsResult, err := svc.GetObjectTagging(tagsInput)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags for object %q, %v", *item.Key, err)
		}

		for _, tag := range tagsResult.TagSet {
			if *tag.Key == "channel" && *tag.Value == channelId {
				// Append the S3 URL for the video
				videos = append(videos, fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, *item.Key))
			}
		}
	}

	return videos, nil
}

func handleChannelVideos(w http.ResponseWriter, r *http.Request) {
	// Extract channelId from the URL path (e.g., /channel/{channelId})
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	channelId := parts[2]

	// Get videos from S3 by channelId
	videos, err := getS3VideosByChannel(channelId)
	if err != nil {
		log.Printf("Failed to retrieve videos for channelId=%s: %v", channelId, err)
		http.Error(w, "Failed to retrieve videos", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Create the response structure
	response := map[string][]string{
		"videos": videos,
	}

	// Encode the response as JSON and send it
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}
