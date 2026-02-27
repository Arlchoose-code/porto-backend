package helpers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type revalidatePayload struct {
	Secret string `json:"secret"`
	Type   string `json:"type"`
	Slug   string `json:"slug,omitempty"`
}

// RevalidateFrontend memanggil Next.js revalidate endpoint
func RevalidateFrontend(revalidateType string, slug string) {
	frontendUrl := os.Getenv("FRONTEND_URL")
	secret := os.Getenv("REVALIDATE_SECRET")

	if frontendUrl == "" || secret == "" {
		return
	}

	payload := revalidatePayload{
		Secret: secret,
		Type:   revalidateType,
		Slug:   slug,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[REVALIDATE ERROR] marshal: %v", err)
		return
	}

	resp, err := http.Post(
		frontendUrl+"/api/revalidate",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		log.Printf("[REVALIDATE ERROR] request: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("[REVALIDATE OK] type=%s slug=%s status=%d", revalidateType, slug, resp.StatusCode)
}
