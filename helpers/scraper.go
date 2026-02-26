package helpers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScrapeArticle mengambil konten artikel dari URL
func ScrapeArticle(articleUrl string) (string, error) {

	// Buat HTTP client dengan timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Buat request dengan user agent biar ga keblok
	req, err := http.NewRequest("GET", articleUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")

	// Kirim request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url: %v", err)
	}
	defer resp.Body.Close()

	// Cek status response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("url returned status %d", resp.StatusCode)
	}

	// Parse HTML dengan goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse html: %v", err)
	}

	// Hapus elemen yang ga relevan
	doc.Find("script, style, nav, header, footer, aside, iframe, noscript, form").Remove()

	// Coba ambil konten dari tag artikel yang umum dipakai
	var content string
	selectors := []string{"article", "main", ".content", ".post-content", ".entry-content", ".article-body", "#content"}

	for _, selector := range selectors {
		text := strings.TrimSpace(doc.Find(selector).Text())
		if len(text) > 200 {
			content = text
			break
		}
	}

	// Fallback ke body kalau ga ketemu
	if content == "" {
		content = strings.TrimSpace(doc.Find("body").Text())
	}

	// Bersihkan whitespace berlebihan
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	content = strings.Join(cleanLines, "\n")

	// Batasi konten maksimal 3000 karakter biar ga overload prompt AI
	if len(content) > 3000 {
		content = content[:3000] + "..."
	}

	return content, nil
}
