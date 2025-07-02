package main

import (
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

func handler(w http.ResponseWriter, r *http.Request) {
	// Match Lightning address pattern
	re := regexp.MustCompile(`^/\.well-known/lnurlp/(\w+)$`)
	matches := re.FindStringSubmatch(r.URL.Path)
	
	if matches == nil {
		// Redirect all other requests to walletofsatoshi.com
		http.Redirect(w, r, "https://walletofsatoshi.com"+r.URL.RequestURI(), http.StatusMovedPermanently)
		return
	}

	username := "gringokiwi"
	
	// Build target URL with query parameters
	targetURL := fmt.Sprintf("https://bipa.app/.well-known/lnurlp/%s", username)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Make request to walletofsatoshi.com
	resp, err := http.Get(targetURL)
	if err != nil {
		log.Printf("Error fetching %s: %v", targetURL, err)
		http.Error(w, "Service temporarily unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers (except connection-related ones)
	for k, v := range resp.Header {
		if !strings.EqualFold(k, "connection") && !strings.EqualFold(k, "content-length") {
			w.Header()[k] = v
		}
	}

	// Set status code and copy body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}