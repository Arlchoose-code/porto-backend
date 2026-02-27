package helpers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/image/webp"
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

// isCompressableImage cek apakah file adalah gambar yang bisa dikompress
func isCompressableImage(ext string) bool {
	compressable := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		".gif": true, ".bmp": true, ".tiff": true, ".tif": true,
	}
	return compressable[ext]
}

// hasTransparencyNRGBA cek apakah NRGBA image punya pixel transparan
func hasTransparencyNRGBA(img *image.NRGBA) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.NRGBAAt(x, y).A < 255 {
				return true
			}
		}
	}
	return false
}

func hasTransparencyRGBA(img *image.RGBA) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.RGBAAt(x, y).A < 255 {
				return true
			}
		}
	}
	return false
}

// toRGBImage konversi image ke RGBA dengan background putih
func toRGBImage(img image.Image) image.Image {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Over)
	return rgba
}

// resizeIfNeeded resize image jika dimensi melebihi maxDim
func resizeIfNeeded(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	if w <= maxDim && h <= maxDim {
		return img
	}

	var newW, newH int
	if w > h {
		newW = maxDim
		newH = int(float64(h) * float64(maxDim) / float64(w))
	} else {
		newH = maxDim
		newW = int(float64(w) * float64(maxDim) / float64(h))
	}
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := bounds.Min.X + x*w/newW
			srcY := bounds.Min.Y + y*h/newH
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}

// compressImageBytes mengompress raw image bytes ke target max size
// Return: compressed bytes, output extension, error
func compressImageBytes(data []byte, origExt string, maxBytes int64) ([]byte, string, error) {
	// Handle WebP secara khusus
	if origExt == ".webp" {
		img, err := webp.Decode(bytes.NewReader(data))
		if err == nil {
			img = resizeIfNeeded(img, 1920)
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, toRGBImage(img), &jpeg.Options{Quality: 85}); err == nil {
				if int64(buf.Len()) < int64(len(data)) {
					return buf.Bytes(), ".jpg", nil
				}
			}
		}
		return data, origExt, nil
	}

	// Decode gambar biasa
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// Tidak bisa decode (misal gif animasi), return asli
		return data, origExt, nil
	}

	// Cek transparansi untuk PNG
	hasTrans := false
	if format == "png" {
		switch typed := img.(type) {
		case *image.NRGBA:
			hasTrans = hasTransparencyNRGBA(typed)
		case *image.RGBA:
			hasTrans = hasTransparencyRGBA(typed)
		}
	}

	// Resize kalau terlalu besar
	img = resizeIfNeeded(img, 1920)

	// PNG transparan → encode PNG tetap
	if hasTrans {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return data, origExt, nil
		}
		if int64(buf.Len()) < int64(len(data)) {
			return buf.Bytes(), ".png", nil
		}
		return data, origExt, nil
	}

	// Semua lainnya → JPEG dengan kualitas adaptif
	quality := 85
	for quality >= 40 {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, toRGBImage(img), &jpeg.Options{Quality: quality}); err != nil {
			break
		}
		if int64(buf.Len()) <= maxBytes || quality == 40 {
			if int64(buf.Len()) < int64(len(data)) {
				return buf.Bytes(), ".jpg", nil
			}
			return data, origExt, nil
		}
		quality -= 10
	}

	return data, origExt, nil
}

// UploadFile menyimpan file ke folder uploads/ dengan kompresi otomatis untuk gambar
func UploadFile(c *gin.Context, fieldName string, folder string) (string, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return "", err
	}

	if err := validateFileExtension(file); err != nil {
		return "", err
	}

	uploadPath := fmt.Sprintf("uploads/%s", folder)
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	timestamp := time.Now().UnixNano()

	// Untuk gambar yang bisa dikompress
	if isCompressableImage(ext) {
		src, err := file.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		var rawBuf bytes.Buffer
		if _, err := rawBuf.ReadFrom(src); err != nil {
			return "", err
		}

		// Kompresi dengan target max 700KB
		compressed, outExt, err := compressImageBytes(rawBuf.Bytes(), ext, 700*1024)
		if err != nil {
			compressed = rawBuf.Bytes()
			outExt = ext
		}

		fileName := fmt.Sprintf("%d%s", timestamp, outExt)
		filePath := fmt.Sprintf("%s/%s", uploadPath, fileName)

		if err := os.WriteFile(filePath, compressed, 0644); err != nil {
			return "", err
		}
		return filePath, nil
	}

	// Non-image atau ico/svg: simpan langsung
	fileName := fmt.Sprintf("%d%s", timestamp, ext)
	filePath := fmt.Sprintf("%s/%s", uploadPath, fileName)
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

	if len(fileUrl) > 4 && fileUrl[:4] == "http" {
		found := false
		for i := 0; i <= len(fileUrl)-8; i++ {
			if fileUrl[i:i+8] == "/uploads" {
				path = fileUrl[i+1:]
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
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".tif":  true,
		".ico":  true,
		".svg":  true,
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
