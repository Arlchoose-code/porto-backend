package helpers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ToolFunc func(input map[string]any) (any, error)

// FieldSchema mendefinisikan satu input field untuk sebuah tool
type FieldSchema struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"`        // "text" | "textarea" | "url" | "number" | "range" | "checkbox" | "select"
	Placeholder string `json:"placeholder"` // untuk text/textarea/url
	Default     any    `json:"default"`     // nilai default
	Min         *int   `json:"min,omitempty"`
	Max         *int   `json:"max,omitempty"`
	Options     []any  `json:"options,omitempty"` // untuk type "select"
	Required    bool   `json:"required"`
}

// DocStep satu langkah di dokumentasi penggunaan tool
type DocStep struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

// ToolMeta metadata sebuah tool: fungsi + schema input + docs
type ToolMeta struct {
	Fn          ToolFunc      `json:"-"`
	Name        string        `json:"name"` // nama readable untuk dropdown dashboard
	InputSchema []FieldSchema `json:"input_schema"`
	Docs        *ToolDocs     `json:"docs,omitempty"` // dokumentasi penggunaan (opsional)
}

// ToolDocs dokumentasi penggunaan tool yang tampil di halaman publik
type ToolDocs struct {
	Description string    `json:"description"`
	Steps       []DocStep `json:"steps,omitempty"`
	Notes       []string  `json:"notes,omitempty"`
	Examples    []string  `json:"examples,omitempty"` // contoh input/output
}

// ToolRegistry — slug → ToolMeta (single source of truth)
var ToolRegistry = map[string]ToolMeta{
	"md5": {
		Fn:   toolMD5,
		Name: "MD5 Hash Generator",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "Text", Type: "textarea", Placeholder: "Text to hash...", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Generate MD5 hash dari teks. MD5 menghasilkan hash 128-bit (32 karakter hex). Cocok untuk checksum file, bukan untuk menyimpan password.",
			Steps: []DocStep{
				{Title: "Masukkan teks", Desc: "Ketik atau paste teks yang ingin di-hash ke input field."},
				{Title: "Klik Run", Desc: "Hash MD5 akan langsung tampil di bawah."},
			},
			Notes:    []string{"MD5 tidak cocok untuk password — gunakan bcrypt/argon2", "Hash yang sama selalu menghasilkan output yang sama (deterministik)", "Tidak bisa di-reverse (one-way)"},
			Examples: []string{"Input: hello → Output: 5d41402abc4b2a76b9719d911017c592"},
		},
	},
	"sha1": {
		Fn:   toolSHA1,
		Name: "SHA-1 Hash Generator",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "Text", Type: "textarea", Placeholder: "Text to hash...", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Generate SHA-1 hash dari teks. Menghasilkan hash 160-bit (40 karakter hex).",
			Steps: []DocStep{
				{Title: "Masukkan teks", Desc: "Ketik atau paste teks ke input field."},
				{Title: "Klik Run", Desc: "Hash SHA-1 akan langsung tampil."},
			},
			Notes:    []string{"SHA-1 sudah dianggap lemah untuk keamanan kriptografi", "Masih umum dipakai untuk Git commit hash dan checksum"},
			Examples: []string{"Input: hello → Output: aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		},
	},
	"sha256": {
		Fn:   toolSHA256,
		Name: "SHA-256 Hash Generator",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "Text", Type: "textarea", Placeholder: "Text to hash...", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Generate SHA-256 hash dari teks. Menghasilkan hash 256-bit (64 karakter hex). Standar industri untuk keamanan data.",
			Steps: []DocStep{
				{Title: "Masukkan teks", Desc: "Ketik atau paste teks ke input field."},
				{Title: "Klik Run", Desc: "Hash SHA-256 akan langsung tampil."},
			},
			Notes:    []string{"SHA-256 adalah standar yang direkomendasikan untuk keamanan", "Digunakan di Bitcoin, SSL/TLS, dan banyak protokol keamanan modern"},
			Examples: []string{"Input: hello → Output: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		},
	},
	"base64-encode": {
		Fn:   toolBase64Encode,
		Name: "Base64 Encode",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "Text", Type: "textarea", Placeholder: "Text to encode...", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Encode teks ke format Base64. Base64 mengkonversi binary data menjadi karakter ASCII yang aman untuk ditransmisikan.",
			Steps: []DocStep{
				{Title: "Masukkan teks", Desc: "Ketik atau paste teks yang ingin di-encode."},
				{Title: "Klik Run", Desc: "Hasil encode Base64 akan tampil di bawah."},
			},
			Notes:    []string{"Base64 bukan enkripsi — siapapun bisa decode-nya", "Ukuran output ~33% lebih besar dari input", "Umum dipakai untuk embed gambar di HTML/CSS dan Basic Auth header"},
			Examples: []string{"Input: Hello World → Output: SGVsbG8gV29ybGQ="},
		},
	},
	"base64-decode": {
		Fn:   toolBase64Decode,
		Name: "Base64 Decode",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "Base64 String", Type: "textarea", Placeholder: "Base64 string to decode...", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Decode string Base64 kembali ke teks aslinya.",
			Steps: []DocStep{
				{Title: "Masukkan Base64 string", Desc: "Paste string Base64 yang ingin di-decode. Pastikan string valid (hanya karakter A-Z, a-z, 0-9, +, /, =)."},
				{Title: "Klik Run", Desc: "Teks asli akan tampil di bawah."},
			},
			Notes:    []string{"Input harus berupa Base64 string yang valid", "Padding = di akhir string boleh ada atau tidak ada"},
			Examples: []string{"Input: SGVsbG8gV29ybGQ= → Output: Hello World"},
		},
	},
	"word-counter": {
		Fn:   toolWordCounter,
		Name: "Word Counter",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "Text", Type: "textarea", Placeholder: "Paste your text here...", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Hitung jumlah kata, karakter, dan baris dari teks. Berguna untuk menulis artikel, essay, atau konten dengan batas kata tertentu.",
			Steps: []DocStep{
				{Title: "Paste teks", Desc: "Copy teks dari mana saja lalu paste ke input field."},
				{Title: "Klik Run", Desc: "Statistik kata, karakter, dan baris akan langsung tampil."},
			},
			Notes: []string{"Kata dihitung berdasarkan spasi dan whitespace", "Characters no space tidak menghitung spasi dan tab", "Line break (enter) dihitung sebagai satu baris"},
		},
	},
	"text-reverse": {
		Fn:   toolTextReverse,
		Name: "Text Reverse",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "Text", Type: "textarea", Placeholder: "Text to reverse...", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Balik urutan karakter dalam teks. Support karakter Unicode termasuk emoji dan karakter non-ASCII.",
			Examples:    []string{"Input: Hello World → Output: dlroW olleH"},
		},
	},
	"uuid-generator": {
		Fn:          toolUUID,
		Name:        "UUID Generator",
		InputSchema: []FieldSchema{},
		Docs: &ToolDocs{
			Description: "Generate UUID v4 secara acak. UUID (Universally Unique Identifier) adalah string 128-bit yang hampir pasti unik di seluruh dunia.",
			Steps: []DocStep{
				{Title: "Klik Generate UUID", Desc: "UUID baru akan dibuat setiap kali tombol diklik."},
				{Title: "Copy hasilnya", Desc: "Klik tombol Copy di sebelah UUID untuk menyalinnya."},
			},
			Notes:    []string{"UUID v4 menggunakan random number — probabilitas duplikat sangat kecil", "Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx", "Cocok untuk primary key database, session ID, atau tracking ID"},
			Examples: []string{"Output contoh: 550e8400-e29b-41d4-a716-446655440000"},
		},
	},
	"password-generator": {
		Fn:   toolPasswordGenerator,
		Name: "Password Generator",
		InputSchema: func() []FieldSchema {
			min4, max128 := 4, 128
			return []FieldSchema{
				{Key: "length", Label: "Length", Type: "range", Default: 16, Min: &min4, Max: &max128},
				{Key: "uppercase", Label: "Uppercase (A-Z)", Type: "checkbox", Default: true},
				{Key: "lowercase", Label: "Lowercase (a-z)", Type: "checkbox", Default: true},
				{Key: "numbers", Label: "Numbers (0-9)", Type: "checkbox", Default: true},
				{Key: "symbols", Label: "Symbols (!@#...)", Type: "checkbox", Default: false},
			}
		}(),
	},
	"json-formatter": {
		Fn:   toolJsonFormatter,
		Name: "JSON Formatter",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "JSON", Type: "textarea", Placeholder: "{\"key\": \"value\"}", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Format JSON yang minified atau tidak rapi menjadi format yang mudah dibaca dengan indentasi 2 spasi.",
			Steps: []DocStep{
				{Title: "Paste JSON", Desc: "Masukkan JSON yang ingin di-format, boleh minified atau tidak rapi."},
				{Title: "Klik Run", Desc: "JSON terformat akan tampil dengan indentasi yang rapi."},
			},
			Notes:    []string{"Input harus berupa JSON yang valid", "Indentasi menggunakan 2 spasi"},
			Examples: []string{`Input: {"name":"John","age":30} → Output terformat dengan indentasi`},
		},
	},
	"json-minifier": {
		Fn:   toolJsonMinifier,
		Name: "JSON Minifier",
		InputSchema: []FieldSchema{
			{Key: "text", Label: "JSON", Type: "textarea", Placeholder: "{\"key\": \"value\"}", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Minify JSON dengan menghapus semua whitespace dan baris kosong yang tidak diperlukan. Berguna untuk mengurangi ukuran payload API.",
			Steps: []DocStep{
				{Title: "Paste JSON", Desc: "Masukkan JSON yang ingin di-minify."},
				{Title: "Klik Run", Desc: "JSON versi compact akan tampil tanpa spasi dan newline."},
			},
			Notes: []string{"Input harus berupa JSON yang valid", "Output lebih kecil tapi sulit dibaca manusia"},
		},
	},
	"http-get": {
		Fn:   toolHttpGet,
		Name: "HTTP GET Request",
		InputSchema: []FieldSchema{
			{Key: "url", Label: "URL", Type: "url", Placeholder: "https://api.example.com/data", Required: true},
		},
		Docs: &ToolDocs{
			Description: "Kirim HTTP GET request ke URL yang ditentukan dan tampilkan response-nya. Berguna untuk test public API atau cek response endpoint.",
			Steps: []DocStep{
				{Title: "Masukkan URL", Desc: "Ketik URL lengkap termasuk https://. URL harus bisa diakses secara publik."},
				{Title: "Klik Send", Desc: "Request akan dikirim dan response body + status code akan tampil."},
			},
			Notes:    []string{"Hanya support GET request", "URL harus public — tidak bisa akses localhost atau private network", "Timeout 10 detik", "Response JSON akan diformat otomatis"},
			Examples: []string{"https://api.github.com/users/github", "https://jsonplaceholder.typicode.com/posts/1"},
		},
	},

	// ==================== GAME NICKNAME CHECKER - DUNIAGAMES ====================
	"check-ign-ml": {
		Fn:   makeDuniagamesCheckIgnFn("MOBILE_LEGENDS"),
		Name: "Mobile Legends: Bang Bang",
		InputSchema: []FieldSchema{
			{Key: "user_id", Label: "User ID", Type: "text", Placeholder: "e.g. 604210151", Required: true},
			{Key: "zone_id", Label: "Zone ID", Type: "text", Placeholder: "e.g. 8425", Required: true},
		},
	},
	"check-ign-ff": {
		Fn:   makeDuniagamesCheckIgnFn("FREEFIRE"),
		Name: "Free Fire",
		InputSchema: []FieldSchema{
			{Key: "user_id", Label: "User ID", Type: "text", Placeholder: "e.g. 116502997", Required: true},
		},
	},
	"check-ign-cod": {
		Fn:   makeDuniagamesCheckIgnFn("CALL_OF_DUTY"),
		Name: "Call of Duty Mobile",
		InputSchema: []FieldSchema{
			{Key: "user_id", Label: "User ID", Type: "text", Placeholder: "e.g. 10808316016143544796", Required: true},
		},
	},
	"check-ign-bloodstrike": {
		Fn:   makeDuniagamesCheckIgnFn("BLOOD_STRIKE"),
		Name: "Blood Strike",
		InputSchema: []FieldSchema{
			{Key: "user_id", Label: "User ID", Type: "text", Placeholder: "e.g. 586027092228", Required: true},
			// Server ID fix -1, gak perlu input dari user
		},
	},
	// ==================== GAME LAINNYA (NANTI TAMBAHIN) ====================
}

// GetRegistrySlugs — daftar semua slug yang terdaftar (untuk API endpoint)
func GetRegistrySlugs() []string {
	slugs := make([]string, 0, len(ToolRegistry))
	for slug := range ToolRegistry {
		slugs = append(slugs, slug)
	}
	return slugs
}

// GetInputSchema — ambil schema input untuk slug tertentu
func GetInputSchema(slug string) ([]FieldSchema, bool) {
	meta, ok := ToolRegistry[slug]
	if !ok {
		return nil, false
	}
	return meta.InputSchema, true
}

// GetDocs — ambil dokumentasi tool untuk slug tertentu
func GetDocs(slug string) *ToolDocs {
	meta, ok := ToolRegistry[slug]
	if !ok {
		return nil
	}
	return meta.Docs
}

// ExecuteTool eksekusi tool berdasarkan slug
func ExecuteTool(slug string, input map[string]any) (any, error) {
	meta, exists := ToolRegistry[slug]
	if !exists {
		return nil, fmt.Errorf("handler for '%s' not found", slug)
	}
	return meta.Fn(input)
}

// ===== TOOL FUNCTIONS (NON-GAME) =====

func toolMD5(input map[string]any) (any, error) {
	text := getString(input, "text")
	h := md5.New()
	h.Write([]byte(text))
	return map[string]any{
		"input":  text,
		"result": fmt.Sprintf("%x", h.Sum(nil)),
	}, nil
}

func toolSHA1(input map[string]any) (any, error) {
	text := getString(input, "text")
	h := sha1.New()
	h.Write([]byte(text))
	return map[string]any{
		"input":  text,
		"result": fmt.Sprintf("%x", h.Sum(nil)),
	}, nil
}

func toolSHA256(input map[string]any) (any, error) {
	text := getString(input, "text")
	h := sha256.New()
	h.Write([]byte(text))
	return map[string]any{
		"input":  text,
		"result": fmt.Sprintf("%x", h.Sum(nil)),
	}, nil
}

func toolBase64Encode(input map[string]any) (any, error) {
	text := getString(input, "text")
	return map[string]any{
		"input":  text,
		"result": base64.StdEncoding.EncodeToString([]byte(text)),
	}, nil
}

func toolBase64Decode(input map[string]any) (any, error) {
	text := getString(input, "text")
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 string")
	}
	return map[string]any{
		"input":  text,
		"result": string(decoded),
	}, nil
}

func toolWordCounter(input map[string]any) (any, error) {
	text := getString(input, "text")
	words := 0
	if strings.TrimSpace(text) != "" {
		words = len(strings.Fields(text))
	}
	return map[string]any{
		"words":               words,
		"characters":          len(text),
		"characters_no_space": len(strings.ReplaceAll(text, " ", "")),
		"lines":               len(strings.Split(text, "\n")),
	}, nil
}

func toolTextReverse(input map[string]any) (any, error) {
	text := getString(input, "text")
	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return map[string]any{
		"input":  text,
		"result": string(runes),
	}, nil
}

func toolUUID(input map[string]any) (any, error) {
	return map[string]any{
		"result": uuid.New().String(),
	}, nil
}

func toolPasswordGenerator(input map[string]any) (any, error) {
	length := getInt(input, "length", 12)
	if length < 4 || length > 128 {
		return nil, fmt.Errorf("length must be between 4 and 128")
	}

	chars := ""
	if getBool(input, "uppercase", true) {
		chars += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if getBool(input, "lowercase", true) {
		chars += "abcdefghijklmnopqrstuvwxyz"
	}
	if getBool(input, "numbers", true) {
		chars += "0123456789"
	}
	if getBool(input, "symbols", false) {
		chars += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}
	if chars == "" {
		chars = "abcdefghijklmnopqrstuvwxyz"
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	password := make([]byte, length)
	for i := range password {
		password[i] = chars[r.Intn(len(chars))]
	}

	return map[string]any{
		"result": string(password),
		"length": length,
	}, nil
}

func toolJsonFormatter(input map[string]any) (any, error) {
	text := getString(input, "text")
	var parsed any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}
	formatted, _ := json.MarshalIndent(parsed, "", "  ")
	return map[string]any{"result": string(formatted)}, nil
}

func toolJsonMinifier(input map[string]any) (any, error) {
	text := getString(input, "text")
	var parsed any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON")
	}
	minified, _ := json.Marshal(parsed)
	return map[string]any{"result": string(minified)}, nil
}

func toolHttpGet(input map[string]any) (any, error) {
	url := getString(input, "url")
	if url == "" {
		return nil, fmt.Errorf("url is required")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return map[string]any{"status": resp.StatusCode, "body": string(body)}, nil
	}
	return map[string]any{"status": resp.StatusCode, "body": parsed}, nil
}

// ===== HELPER =====

func getString(input map[string]any, key string) string {
	if val, ok := input[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func getInt(input map[string]any, key string, defaultVal int) int {
	if val, ok := input[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			var i int
			fmt.Sscanf(v, "%d", &i)
			return i
		}
	}
	return defaultVal
}

func getBool(input map[string]any, key string, defaultVal bool) bool {
	if val, ok := input[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "1"
		}
	}
	return defaultVal
}

// ==================== DUNIAGAMES IMPLEMENTATION ====================

const duniagamesAPI = "https://api.duniagames.co.id/api/transaction/v1/top-up/inquiry/store"

type duniagamesGameMeta struct {
	ProductId  int
	ItemId     int
	CatalogId  int
	PaymentId  int
	ProductRef string
	Denom      string
	NeedZone   bool
}

var duniagamesGames = map[string]duniagamesGameMeta{
	"MOBILE_LEGENDS": {
		ProductId:  1,
		ItemId:     3,
		CatalogId:  58,
		PaymentId:  353,
		ProductRef: "CMS",
		Denom:      "REG",
		NeedZone:   true,
	},
	"FREEFIRE": {
		ProductId:  3,
		ItemId:     353,
		CatalogId:  376,
		PaymentId:  1252,
		ProductRef: "CMS",
		Denom:      "REG",
		NeedZone:   false,
	},
	"CALL_OF_DUTY": {
		ProductId:  18,
		ItemId:     88,
		CatalogId:  144,
		PaymentId:  828,
		ProductRef: "CMS",
		Denom:      "REG",
		NeedZone:   false,
	},
	"BLOOD_STRIKE": {
		ProductId:  149,
		ItemId:     1654,
		CatalogId:  2836,
		PaymentId:  7578,
		ProductRef: "REG",
		Denom:      "REG",
		NeedZone:   false, // Pake serverId, tapi kita handle khusus
	},
}

// makeDuniagamesCheckIgnFn — factory untuk pake API DuniaGames
func makeDuniagamesCheckIgnFn(game string) ToolFunc {
	return func(input map[string]any) (any, error) {
		userID := getString(input, "user_id")
		zoneID := getString(input, "zone_id")

		if userID == "" {
			return nil, fmt.Errorf("user_id wajib diisi")
		}

		meta, ok := duniagamesGames[game]
		if !ok {
			return nil, fmt.Errorf("game '%s' gak ditemukan di Duniagames", game)
		}

		// Cek zone_id kalo emang butuh
		if meta.NeedZone && zoneID == "" {
			return nil, fmt.Errorf("zone_id wajib diisi untuk game ini")
		}

		// Build payload
		// Build payload
		payload := map[string]any{
			"productId":         meta.ProductId,
			"itemId":            meta.ItemId,
			"catalogId":         meta.CatalogId,
			"paymentId":         meta.PaymentId,
			"gameId":            userID,
			"product_ref":       meta.ProductRef,
			"product_ref_denom": meta.Denom,
		}

		// Khusus Blood Strike pake serverId
		if game == "BLOOD_STRIKE" {
			payload["serverId"] = "-1"
			payload["serverName"] = "Blood Strike"
		} else if meta.NeedZone {
			payload["zoneId"] = zoneID
		}

		payloadBytes, _ := json.Marshal(payload)

		// Request ke DuniaGames
		client := &http.Client{Timeout: 15 * time.Second}
		req, err := http.NewRequest("POST", duniagamesAPI, strings.NewReader(string(payloadBytes)))
		if err != nil {
			return nil, fmt.Errorf("gagal bikin request: %v", err)
		}

		req.Header.Set("Host", "api.duniagames.co.id")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:83.0) Gecko/20100101 Firefox/83.0")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request gagal: %v", err)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)

		var result map[string]any
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("response Duniagames error: %v", err)
		}

		// Cek response
		if code, ok := result["statusCode"].(float64); ok && code != 200 {
			msg, _ := result["message"].(string)
			return nil, fmt.Errorf("duniagames error: %s (code: %.0f)", msg, code)
		}

		// Ambil data
		data, ok := result["data"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("struktur response salah")
		}

		nickname, ok := data["userNameGame"].(string)
		if !ok || nickname == "" {
			return nil, fmt.Errorf("nickname gak ketemu")
		}

		out := map[string]any{
			"nickname": nickname,
			"game":     game,
			"user_id":  userID,
		}
		if zoneID != "" {
			out["zone_id"] = zoneID
		}
		return out, nil
	}
}
