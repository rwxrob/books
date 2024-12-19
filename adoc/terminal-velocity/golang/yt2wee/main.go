package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	apiKey          = "YOUR_YOUTUBE_API_KEY" // Replace with your API key
	channelID       = "YOUR_CHANNEL_ID"      // Replace with your channel ID
	pollInterval    = 5 * time.Second        // Adjust as needed
	expiryTime      = 10 * time.Minute       // Time after which entries expire
	cleanupInterval = 1 * time.Minute        // Interval to run cleanup
)

// ChatMessage represents a chat message from the source
type ChatMessage struct {
	User    string
	Message string
}

// Global map with timestamps and a mutex for thread safety
var (
	seenMessages = make(map[string]time.Time)
	mu           sync.Mutex
)

// fetchChatMessages fetches messages from YouTube Live Chat and deduplicates them
func fetchChatMessages(service *youtube.Service, liveChatID string, nextPageToken string) (messages []ChatMessage, newNextPageToken string, err error) {
	call := service.LiveChatMessages.List(liveChatID, []string{"snippet", "authorDetails"}).Key(apiKey)
	if nextPageToken != "" {
		call = call.PageToken(nextPageToken)
	}

	response, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("error fetching live chat messages: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	for _, item := range response.Items {
		if _, exists := seenMessages[item.Id]; !exists {
			seenMessages[item.Id] = now // Mark the message as seen with the current timestamp
			messages = append(messages, ChatMessage{
				User:    item.AuthorDetails.DisplayName,
				Message: item.Snippet.DisplayMessage,
			})
		}
	}

	return messages, response.NextPageToken, nil
}

// cleanupSeenMessages removes old entries from the seenMessages map
func cleanupSeenMessages() {
	for {
		time.Sleep(cleanupInterval)

		mu.Lock()
		now := time.Now()
		for id, timestamp := range seenMessages {
			if now.Sub(timestamp) > expiryTime {
				delete(seenMessages, id)
			}
		}
		mu.Unlock()
	}
}

// getLiveChatID fetches the live chat ID for the current live stream
func getLiveChatID(service *youtube.Service) (string, error) {
	call := service.Search.List([]string{"snippet"}).
		ChannelId(channelID).
		EventType("live").
		Type("video").
		Key(apiKey)

	response, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("error fetching live broadcasts: %w", err)
	}

	if len(response.Items) == 0 {
		return "", fmt.Errorf("no live broadcasts found for channel ID %s", channelID)
	}

	videoID := response.Items[0].Id.VideoId

	videoCall := service.Videos.List([]string{"liveStreamingDetails"}).
		Id(videoID).
		Key(apiKey)

	videoResponse, err := videoCall.Do()
	if err != nil {
		return "", fmt.Errorf("error fetching video details: %w", err)
	}

	if len(videoResponse.Items) == 0 {
		return "", fmt.Errorf("no video details found for video ID %s", videoID)
	}

	liveChatID := videoResponse.Items[0].LiveStreamingDetails.ActiveLiveChatId
	if liveChatID == "" {
		return "", fmt.Errorf("live chat ID not found for video ID %s", videoID)
	}

	return liveChatID, nil
}

func main() {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating YouTube service: %v", err)
	}

	liveChatID, err := getLiveChatID(service)
	if err != nil {
		log.Fatalf("Error getting live chat ID: %v", err)
	}

	// Start the cleanup routine
	go cleanupSeenMessages()

	var nextPageToken string
	for {
		messages, newNextPageToken, err := fetchChatMessages(service, liveChatID, nextPageToken)
		if err != nil {
			log.Printf("Error fetching chat messages: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		for _, msg := range messages {
			fmt.Printf("%s: %s\n", msg.User, msg.Message)
			// Here, you would send the message to WeeChat FIFO
		}

		nextPageToken = newNextPageToken
		time.Sleep(pollInterval)
	}
}
