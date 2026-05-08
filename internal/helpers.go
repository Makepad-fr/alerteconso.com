package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// getQueryInt extracts an int query parameter with a default fallback
func getQueryInt(r *http.Request, key string, defaultValue int) int {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultValue
	}
	if val, err := strconv.Atoi(valStr); err == nil && val > 0 {
		return val
	}
	return defaultValue
}

// max returns the maximum of two ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func containsPipe(s string) bool {
	return len(s) > 0 && len(splitOnPipe(s)) > 0
}

func splitOnPipe(s string) []string {
	parts := strings.Split(s, "|")
	var urls []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			urls = append(urls, trimmed)
		}
	}
	return urls
}
func parseImageURLs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	// 1. Try parsing as JSON array
	var urls []string
	err := json.Unmarshal([]byte(raw), &urls)
	if err == nil && len(urls) > 0 {
		return urls
	}

	// 2. Try quoted string
	var single string
	quoted := `"` + strings.ReplaceAll(raw, `"`, `\"`) + `"`
	if err2 := json.Unmarshal([]byte(quoted), &single); err2 == nil {
		if strings.Contains(single, "|") {
			return splitOnPipe(single)
		}
		return []string{single}
	}

	// 3. Try raw pipe-delimited string
	if strings.Contains(raw, "|") {
		return splitOnPipe(raw)
	}

	log.Printf("⚠️ Could not parse image field: %s", raw)
	return nil
}

// define your helper functions
func minus(a int) int {
	return a - 1
}

func add(a int) int {
	return a + 1
}
