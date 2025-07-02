package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", handler)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func editLnurlpJson(body []byte, username string) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return body
	}

	// Replace metadata with custom value
	data["metadata"] = `[["text/plain","Pay to Wallet of Satoshi user: gringokiwi"],["text/identifier","gringokiwi@walletofsatoshi.com"]]`

	modified, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return body
	}

	return modified
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Handle preflight OPTIONS requests
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Match Lightning Address URL pattern
	re := regexp.MustCompile(`^/\.well-known/lnurlp/(\w+)$`)
	matches := re.FindStringSubmatch(r.URL.Path)
	
	// Redirect if no match found
	if matches == nil {
		http.Redirect(w, r, "https://walletofsatoshi.com"+r.URL.RequestURI(), http.StatusMovedPermanently)
		return
	}

	// Extract username
	username := matches[1]

	// Redirect to walletofsatoshi.com -- unless it's 'gringokiwi'
	targetURL := fmt.Sprintf("https://walletofsatoshi.com/.well-known/lnurlp/%s", username)
	if username == "gringokiwi" {
		targetURL = "https://bipa.app/.well-known/lnurlp/gringokiwi"
	}

	// Append query parameters if present
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Fetch LNURLP JSON from target URL
	resp, err := http.Get(targetURL)
	if err != nil {
		log.Printf("Error fetching %s: %v", targetURL, err)
		http.Error(w, "Service temporarily unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		http.Error(w, "Service temporarily unavailable", http.StatusBadGateway)
		return
	}

	// POC: Modify LNURLP JSON for 'gringokiwi'
	if username == "gringokiwi" {
		body = editLnurlpJson(body, username)
	}

	// Copy response headers (except connection-related ones)
	for k, v := range resp.Header {
		if !strings.EqualFold(k, "connection") && !strings.EqualFold(k, "content-length") {
			w.Header()[k] = v
		}
	}

	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Set status code and return body
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}