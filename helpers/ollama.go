package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

// getOllamaModel mengambil model ollama dari env, default ke llama3
func getOllamaModel() string {
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		return "llama3"
	}
	return model
}

// getOllamaUrl mengambil URL ollama dari env, default ke localhost
func getOllamaUrl() string {
	ollamaUrl := os.Getenv("OLLAMA_URL")
	if ollamaUrl == "" {
		return "http://localhost:11434"
	}
	return ollamaUrl
}

// askOllama mengirim prompt ke Ollama dan mengembalikan response
func askOllama(prompt string) (string, error) {

	reqBody := OllamaRequest{
		Model:  getOllamaModel(),
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(
		getOllamaUrl()+"/api/generate",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect to ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %v", err)
	}

	return ollamaResp.Response, nil
}

// GenerateBlogTitles meminta Ollama untuk generate judul-judul blog
// GenerateBlogTitles meminta Ollama untuk generate judul-judul blog
func GenerateBlogTitles(keyword string, total int) ([]string, error) {

	now := time.Now()
	currentDate := now.Format("2 January 2006")

	var prompt string
	if keyword != "" {
		prompt = fmt.Sprintf(`Kamu adalah editor blog profesional Indonesia.
Hari ini tanggal %s.

Buat %d judul artikel blog dalam Bahasa Indonesia berdasarkan topik: "%s"

ATURAN KETAT:
- Judul harus REALISTIS dan FAKTUAL, bukan fiksi atau spekulasi liar
- Fokus pada fakta, berita, analisis, atau panduan praktis
- JANGAN buat judul tentang skenario fiktif (seperti "di luar angkasa", "di masa depan 2050", dll)
- JANGAN buat judul tentang event yang sudah selesai di masa lalu
- Judul harus bisa dicari di internet dan punya referensi nyata
- Singkat, jelas, SEO-friendly, maksimal 10 kata per judul
- JANGAN tambahkan nomor, tanda strip, atau penjelasan

Balas HANYA daftar judul, satu per baris.`, currentDate, total, keyword)
	} else {
		prompt = fmt.Sprintf(`Kamu adalah editor blog teknologi profesional Indonesia.
Hari ini tanggal %s.

Buat %d judul artikel blog teknologi terkini dalam Bahasa Indonesia.

ATURAN KETAT:
- Topik: AI, cloud computing, programming, cybersecurity, startup Indonesia, mobile dev
- Judul harus REALISTIS, faktual, bisa dicari referensinya
- JANGAN buat judul fiksi atau spekulasi liar
- Singkat, jelas, SEO-friendly, maksimal 10 kata per judul
- JANGAN tambahkan nomor, tanda strip, atau penjelasan

Balas HANYA daftar judul, satu per baris.`, currentDate, total)
	}

	response, err := askOllama(prompt)
	if err != nil {
		return nil, err
	}

	var titles []string
	lines := SplitLines(response)
	for _, line := range lines {
		line = CleanLine(line)
		if line != "" && len(line) > 10 {
			titles = append(titles, line)
		}
	}

	if len(titles) > total {
		titles = titles[:total]
	}

	return titles, nil
}

// GenerateBlogContent meminta Ollama untuk menulis blog dari referensi artikel
func GenerateBlogContent(title string, references []string) (string, string, error) {

	// Gabungkan semua referensi
	var refText string
	for i, ref := range references {
		refText += fmt.Sprintf("=== Referensi %d ===\n%s\n\n", i+1, ref)
	}

	prompt := fmt.Sprintf(`Kamu adalah Aibys, AI Assistant dari Arlchoose yang bertugas menulis artikel blog teknologi dalam Bahasa Indonesia.

Judul artikel yang harus kamu tulis: "%s"

Berikut adalah referensi artikel yang bisa kamu gunakan sebagai sumber informasi:
%s

Instruksi penulisan:
- Tulis artikel yang informatif dan menarik dalam Bahasa Indonesia
- JANGAN menyalin atau memparafrase referensi secara langsung, tulis dengan gaya dan perspektifmu sendiri
- Gunakan informasi dari referensi sebagai dasar fakta, tapi sampaikan dengan cara yang unik
- Format artikel menggunakan HTML (gunakan tag h2, h3, p, ul, li, strong, em)
- Panjang artikel minimal 500 kata
- Sertakan intro yang menarik dan kesimpulan yang berkesan
- Tulis deskripsi singkat (1-2 kalimat) di awal sebelum konten HTML, pisahkan dengan tanda "---DESCRIPTION---" dan "---CONTENT---"

Format response:
---DESCRIPTION---
[deskripsi singkat artikel]
---CONTENT---
[konten artikel dalam HTML]`, title, refText)

	response, err := askOllama(prompt)
	if err != nil {
		return "", "", err
	}

	// Parse description dan content dari response
	description, content := parseOllamaResponse(response)

	return description, content, nil
}

// AskOllama expose askOllama ke package lain
func AskOllama(prompt string) (string, error) {
	return askOllama(prompt)
}

// SplitLines memisahkan string jadi slice of lines
func SplitLines(s string) []string {
	var lines []string
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// CleanLine membersihkan whitespace dan karakter tidak perlu
func CleanLine(s string) string {
	s = TrimSpace(s)
	// Hapus prefix nomor seperti "1. " atau "- "
	for len(s) > 2 && (s[0] == '-' || (s[0] >= '0' && s[0] <= '9')) {
		if s[1] == '.' || s[1] == ')' || s[1] == ' ' {
			s = TrimSpace(s[2:])
		} else {
			break
		}
	}
	return s
}

// TrimSpace menghapus whitespace di awal dan akhir string
func TrimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// parseOllamaResponse memisahkan description dan content dari response Ollama
func parseOllamaResponse(response string) (string, string) {
	description := ""
	content := response

	if idx := indexOfStr(response, "---DESCRIPTION---"); idx != -1 {
		afterDesc := response[idx+len("---DESCRIPTION---"):]
		if idx2 := indexOfStr(afterDesc, "---CONTENT---"); idx2 != -1 {
			description = TrimSpace(afterDesc[:idx2])
			content = afterDesc[idx2+len("---CONTENT---"):]
		}
	}

	return TrimSpace(description), TrimSpace(content)
}

// ParseOllamaResponse expose parseOllamaResponse ke package lain
func ParseOllamaResponse(response string) (string, string) {
	return parseOllamaResponse(response)
}

// indexOfStr mencari index substring dalam string
func indexOfStr(s, substr string) int {
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// CleanAIOutput membersihkan output AI dari karakter markdown yang tidak diinginkan
func CleanAIOutput(s string) string {
	// Hapus opening/closing backtick blocks
	s = strings.ReplaceAll(s, "```html", "")
	s = strings.ReplaceAll(s, "```HTML", "")
	s = strings.ReplaceAll(s, "```", "")
	return TrimSpace(s)
}
