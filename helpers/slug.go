package helpers

import (
	"regexp"
	"strings"
)

// GenerateSlug mengubah string title menjadi slug URL-friendly
// contoh: "My First Project!" → "my-first-project"
func GenerateSlug(title string) string {

	// Ubah ke lowercase
	slug := strings.ToLower(title)

	// Ganti semua karakter selain huruf, angka, dan spasi dengan kosong
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = reg.ReplaceAllString(slug, "")

	// Ganti spasi dan strip berulang dengan single strip
	reg = regexp.MustCompile(`[\s-]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim strip di awal dan akhir
	slug = strings.Trim(slug, "-")

	return slug
}
