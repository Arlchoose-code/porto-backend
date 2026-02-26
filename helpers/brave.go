package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type BraveSearchResult struct {
	Title       string `json:"title"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

type BraveSearchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			Url         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

// cleanQuery hanya bersihkan karakter bermasalah, TIDAK memotong kata
func cleanQuery(title string) string {
	replacer := strings.NewReplacer(
		"\"", "", "'", "", "\\", "", "#", "",
		":", "", "?", "", "!", "",
	)
	clean := strings.TrimSpace(replacer.Replace(title))

	// Batasi maksimal 100 karakter untuk hindari 422
	if len(clean) > 100 {
		// Potong di kata terakhir yang masuk dalam 100 karakter
		words := strings.Fields(clean)
		result := ""
		for _, w := range words {
			if len(result)+len(w)+1 > 100 {
				break
			}
			if result != "" {
				result += " "
			}
			result += w
		}
		return result
	}
	return clean
}

// SearchBrave mencari artikel menggunakan Brave Search API
func SearchBrave(query string) ([]BraveSearchResult, error) {

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("BRAVE_API_KEY is not set")
	}

	// Bersihkan query tapi JANGAN dipotong
	cleanedQuery := cleanQuery(query)
	log.Printf("[BRAVE QUERY] %s", cleanedQuery)

	// Delay hindari rate limit
	time.Sleep(1 * time.Second)

	// count=2 untuk dapat lebih banyak referensi
	searchUrl := fmt.Sprintf(
		"https://api.search.brave.com/res/v1/web/search?q=%s&count=2&search_lang=ms",
		url.QueryEscape(cleanedQuery),
	)

	log.Printf("[BRAVE URL] %s", searchUrl)

	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search brave: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("[BRAVE STATUS] %d", resp.StatusCode)

	if resp.StatusCode == 422 {
		return nil, fmt.Errorf("brave api 422 - invalid request")
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("brave api 429 - rate limit exceeded")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brave api returned status %d", resp.StatusCode)
	}

	var braveResp BraveSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&braveResp); err != nil {
		return nil, fmt.Errorf("failed to decode brave response: %v", err)
	}

	var results []BraveSearchResult
	for _, r := range braveResp.Web.Results {
		results = append(results, BraveSearchResult{
			Title:       r.Title,
			Url:         r.Url,
			Description: r.Description,
		})
	}

	log.Printf("[BRAVE RESULTS] found %d results", len(results))

	return results, nil
}
