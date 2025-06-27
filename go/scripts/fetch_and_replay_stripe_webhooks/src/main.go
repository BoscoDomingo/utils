package main

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	STRIPE_BASE_URL = "https://api.stripe.com/v1"
	DATA_FILE       = "./data/failed_payloads.json"
)

// Simple event struct for ID extraction
type Event struct {
	ID string `json:"id"`
}

// Load environment variables from .env file
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return // .env file is optional
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Only set if not already set
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

// Fetch failed webhook events from Stripe
func fetchFailedEvents() error {
	apiKey := os.Getenv("STRIPE_SECRET")
	if apiKey == "" {
		return fmt.Errorf("STRIPE_SECRET environment variable is required")
	}

	client := &http.Client{}
	var allEvents []json.RawMessage
	startingAfter := ""

	for {
		// Build URL with parameters
		u, _ := url.Parse(STRIPE_BASE_URL + "/events")
		q := u.Query()
		q.Set("limit", "100")
		q.Set("delivery_success", "false")
		if startingAfter != "" {
			q.Set("starting_after", startingAfter)
		}
		u.RawQuery = q.Encode()

		// Create request with auth
		req, _ := http.NewRequest("GET", u.String(), nil)
		auth := base64.StdEncoding.EncodeToString([]byte(apiKey + ":"))
		req.Header.Set("Authorization", "Basic "+auth)

		// Make request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("API error: %d %s", resp.StatusCode, string(body))
		}

		// Parse response
		var result struct {
			Data    []json.RawMessage `json:"data"`
			HasMore bool              `json:"has_more"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		allEvents = append(allEvents, result.Data...)
		if !result.HasMore {
			break
		}

		// Get last event ID for pagination
		var lastEvent Event
		json.Unmarshal(result.Data[len(result.Data)-1], &lastEvent)
		startingAfter = lastEvent.ID
	}

	// Create data directory if it doesn't exist
	os.MkdirAll("./data", 0755)

	// Save to file
	file, err := os.Create(DATA_FILE)
	if err != nil {
		return fmt.Errorf("create file error: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allEvents); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	fmt.Printf("Saved %d failed events to %s\n", len(allEvents), DATA_FILE)
	return nil
}

// Generate Stripe webhook signature
func generateSignature(payload, secret string) string {
	timestamp := time.Now().Unix()
	signedPayload := fmt.Sprintf("%d.%s", timestamp, payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signedPayload))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", timestamp, signature)
}

// Replay webhook events
func replayEvents() error {
	signingSecret := os.Getenv("SIGNING_SECRET")
	if signingSecret == "" {
		return fmt.Errorf("SIGNING_SECRET environment variable is required")
	}

	endpointURL := os.Getenv("ENDPOINT_URL")
	if endpointURL == "" {
		return fmt.Errorf("ENDPOINT_URL environment variable is required")
	}

	// Read events from file
	file, err := os.Open(DATA_FILE)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer file.Close()

	var events []json.RawMessage
	if err := json.NewDecoder(file).Decode(&events); err != nil {
		return fmt.Errorf("parse JSON error: %w", err)
	}

	// Track stats and replayed events
	var success, failed, skipped int
	replayed := make(map[string]bool)
	client := &http.Client{}

	for _, eventData := range events {
		var event Event
		if err := json.Unmarshal(eventData, &event); err != nil {
			fmt.Printf("❌ Failed to parse event: %v\n", err)
			failed++
			continue
		}

		// Skip duplicates
		if replayed[event.ID] {
			fmt.Printf("⏭️  Skipping already replayed event %s\n", event.ID)
			skipped++
			continue
		}

		// Create and send request
		payload := string(eventData)
		signature := generateSignature(payload, signingSecret)

		req, _ := http.NewRequest("POST", endpointURL, bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Stripe-Signature", signature)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Failed event %s: %v\n", event.ID, err)
			failed++
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			replayed[event.ID] = true
			success++
			fmt.Printf("✔️  Replayed event %s\n", event.ID)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("❌ Failed event %s: %d %s\n", event.ID, resp.StatusCode, string(body))
			failed++
		}
		resp.Body.Close()
	}

	// Print summary
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   ✔️ Successfully replayed: %d events\n", success)
	fmt.Printf("   ❌ Failed: %d events\n", failed)
	fmt.Printf("   ⏭️ Skipped (duplicates): %d events\n", skipped)
	fmt.Printf("   📝 Total processed: %d events\n", len(events))

	return nil
}

func main() {
	loadEnv()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <fetch|replay|fetch-and-replay>\n", os.Args[0])
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "fetch":
		err = fetchFailedEvents()
	case "replay":
		err = replayEvents()
	case "fetch-and-replay":
		fmt.Println("🔄 Fetching failed events...")
		if err = fetchFailedEvents(); err != nil {
			break
		}
		fmt.Println("\n🚀 Starting replay...")
		err = replayEvents()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: %s <fetch|replay|fetch-and-replay>\n", os.Args[1], os.Args[0])
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}