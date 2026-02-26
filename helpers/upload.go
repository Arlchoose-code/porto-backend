package helpers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetBaseUrl mengambil base URL dari env
func GetBaseUrl() string {
	baseUrl := os.Getenv("APP_URL")
	if baseUrl == "" {
		return "http://localhost:3000"
	}
	return baseUrl
}

// GetFileUrl mengubah path lokal jadi relative URL
func GetFileUrl(path string) string {
	if path == "" {
		return ""
	}
	return "/" + path
}

// UploadFile menyimpan file ke folder uploads/ dan mengembalikan path-nya
func UploadFile(c *gin.Context, fieldName string, folder string) (string, error) {

	// Ambil file dari request
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", err
	}

	// Validasi ekstensi file
	if err := validateFileExtension(file); err != nil {
		return "", err
	}

	// Buat folder jika belum ada
	uploadPath := fmt.Sprintf("uploads/%s", folder)
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return "", err
	}

	// Generate nama file unik pakai timestamp
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := fmt.Sprintf("%s/%s", uploadPath, fileName)

	// Simpan file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func DeleteFile(fileUrl string) error {
	if fileUrl == "" {
		return nil
	}

	path := fileUrl

	// Kalau full URL → ekstrak path relatifnya
	// http://localhost:3000/uploads/blogs/xxx.png → uploads/blogs/xxx.png
	if len(fileUrl) > 4 && fileUrl[:4] == "http" {
		// Cari "/uploads" dalam URL
		found := false
		for i := 0; i <= len(fileUrl)-8; i++ {
			if fileUrl[i:i+8] == "/uploads" {
				path = fileUrl[i+1:] // hapus leading slash
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return nil
}

// validateFileExtension memvalidasi ekstensi file yang diizinkan
func validateFileExtension(file *multipart.FileHeader) error {
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}

	return nil
}
