package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// SSE manager
var (
	sseClients = make(map[chan string]bool)
	sseMutex   sync.Mutex
)

func addSSEClient(ch chan string) {
	sseMutex.Lock()
	sseClients[ch] = true
	sseMutex.Unlock()
}

func removeSSEClient(ch chan string) {
	sseMutex.Lock()
	delete(sseClients, ch)
	sseMutex.Unlock()
}

func broadcastSSE(msg string) {
	sseMutex.Lock()
	for ch := range sseClients {
		select {
		case ch <- msg:
		default:
		}
	}
	sseMutex.Unlock()
}

// GET /api/blogs/stream — SSE endpoint
func BlogStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	ch := make(chan string, 10)
	addSSEClient(ch)
	defer removeSSEClient(ch)

	notify := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg := <-ch:
			c.SSEvent("message", msg)
			return true
		case <-notify:
			return false
		}
	})
}

func assignTagsToBlog(blog *models.Blog) {
	prompt := fmt.Sprintf(`Berikan 3-5 tag yang relevan untuk artikel berjudul: "%s"
Balas HANYA dengan nama tag, satu per baris, huruf kecil, tanpa penjelasan.
Contoh:
golang
backend
tutorial`, blog.Title)

	response, err := helpers.AskOllama(prompt)
	if err != nil {
		log.Printf("[TAGS ERROR] %v", err)
		return
	}

	lines := helpers.SplitLines(response)
	var tags []models.Tag

	for _, line := range lines {
		tagName := helpers.CleanLine(line)
		tagName = strings.Trim(tagName, "`*_")
		if tagName == "" || len(tagName) < 2 {
			continue
		}

		var tag models.Tag
		tagSlug := helpers.GenerateSlug(tagName)

		if err := database.DB.Where("slug = ?", tagSlug).First(&tag).Error; err != nil {
			tag = models.Tag{Name: tagName, Slug: tagSlug}
			if err := database.DB.Create(&tag).Error; err != nil {
				log.Printf("[TAGS ERROR] failed to create tag: %s, err: %v", tagName, err)
				continue
			}
		}
		tags = append(tags, tag)
	}

	if len(tags) > 0 {
		database.DB.Model(blog).Association("Tags").Replace(tags)
		log.Printf("[TAGS OK] assigned %d tags to blog: %s", len(tags), blog.Title)
	}
}

// POST /api/blogs/generate
func GenerateAiBlog(c *gin.Context) {

	var req structs.AiBlogGenerateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	log.Printf("[GENERATE] start background: keyword=%s, total=%d", req.Keyword, req.Total)

	go generateBlogsBackground(req.Keyword, req.Total)

	c.JSON(http.StatusAccepted, structs.SuccessResponse{
		Success: true,
		Message: "Blog generation started in background",
		Data: map[string]any{
			"keyword": req.Keyword,
			"total":   req.Total,
			"status":  "processing",
		},
	})
}

func generateBlogsBackground(keyword string, total int) {

	log.Printf("[BG] start: keyword=%s, total=%d", keyword, total)

	// Broadcast: mulai generate
	broadcastSSE(fmt.Sprintf(
		`{"type":"generate_progress","keyword":"%s","saved":0,"total_target":%d,"status":"generating_titles"}`,
		keyword, total,
	))

	titles, err := helpers.GenerateBlogTitles(keyword, total)
	if err != nil {
		log.Printf("[BG ERROR] generate titles: %v", err)
		broadcastSSE(fmt.Sprintf(
			`{"type":"generate_done","keyword":"%s","saved":0,"total_target":%d,"failed":true}`,
			keyword, total,
		))
		return
	}

	log.Printf("[BG] titles: %v", titles)

	type titleWithRefs struct {
		Title      string
		References []string
	}

	var titlesWithRefs []titleWithRefs

	for _, title := range titles {
		searchResults, err := helpers.SearchBrave(title)
		if err != nil {
			log.Printf("[BG BRAVE ERROR] %s: %v", title, err)
			titlesWithRefs = append(titlesWithRefs, titleWithRefs{Title: title})
			continue
		}

		var refs []string
		for _, r := range searchResults {
			if r.Description != "" {
				refs = append(refs, r.Description)
			}
			refs = append(refs, r.Url)
		}
		titlesWithRefs = append(titlesWithRefs, titleWithRefs{Title: title, References: refs})
	}

	// Scrape URLs
	for i, twr := range titlesWithRefs {
		var scrapedRefs []string
		for _, ref := range twr.References {
			if len(ref) > 4 && ref[:4] == "http" {
				content, err := helpers.ScrapeArticle(ref)
				if err != nil || content == "" {
					continue
				}
				scrapedRefs = append(scrapedRefs, content)
			} else {
				scrapedRefs = append(scrapedRefs, ref)
			}
		}
		titlesWithRefs[i].References = scrapedRefs
	}

	// Generate konten — broadcast progress per blog
	savedCount := 0
	totalTarget := len(titlesWithRefs)

	for idx, twr := range titlesWithRefs {
		// Broadcast: sedang nulis blog ke-N
		broadcastSSE(fmt.Sprintf(
			`{"type":"generate_progress","keyword":"%s","saved":%d,"total_target":%d,"current_title":"%s","status":"writing"}`,
			keyword, savedCount, totalTarget,
			strings.ReplaceAll(twr.Title, `"`, `'`),
		))

		if len(twr.References) == 0 {
			log.Printf("[BG SKIP] no refs: %s", twr.Title)
			continue
		}

		description, content, err := helpers.GenerateBlogContent(twr.Title, twr.References)
		if err != nil {
			log.Printf("[BG CONTENT ERROR] %s: %v", twr.Title, err)
			continue
		}

		content = helpers.CleanAIOutput(content)
		description = helpers.CleanAIOutput(description)

		blog := models.Blog{
			Title:       twr.Title,
			Slug:        helpers.GenerateSlug(twr.Title),
			Description: description,
			Content:     content,
			Author:      "aibys",
			Status:      "pending",
		}

		if err := database.DB.Create(&blog).Error; err != nil {
			log.Printf("[BG DB ERROR] %s: %v", twr.Title, err)
			continue
		}

		assignTagsToBlog(&blog)
		savedCount++
		log.Printf("[BG OK] saved %d/%d: %s", savedCount, totalTarget, blog.Title)

		// Broadcast: blog ke-N selesai disimpan
		broadcastSSE(fmt.Sprintf(
			`{"type":"generate_progress","keyword":"%s","saved":%d,"total_target":%d,"current_title":"%s","status":"saved"}`,
			keyword, savedCount, totalTarget,
			strings.ReplaceAll(twr.Title, `"`, `'`),
		))

		_ = idx
	}

	log.Printf("[BG] done: keyword=%s, saved: %d/%d", keyword, savedCount, totalTarget)

	// Broadcast: selesai semua
	broadcastSSE(fmt.Sprintf(
		`{"type":"generate_done","keyword":"%s","saved":%d,"total_target":%d,"failed":%t}`,
		keyword, savedCount, totalTarget, savedCount == 0,
	))
}

// PUT /api/blogs/:id/publish
func PublishBlog(c *gin.Context) {
	id := c.Param("id")
	var blog models.Blog

	if err := database.DB.Preload("Tags").Preload("User").First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Blog not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	blog.Status = "published"
	blog.RejectComment = ""

	if err := database.DB.Save(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to publish blog",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Blog published successfully",
		Data:    blog,
	})
}

// PUT /api/blogs/:id/reject
func RejectBlog(c *gin.Context) {
	id := c.Param("id")
	var blog models.Blog

	if err := database.DB.Preload("Tags").First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Blog not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	var req structs.BlogRejectRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	blog.Status = "rejected"
	blog.RejectComment = req.Comment

	if err := database.DB.Save(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to reject blog",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if blog.Author == "aibys" {
		blogCopy := blog
		go regenerateBlog(&blogCopy, req.Comment)

		c.JSON(http.StatusOK, structs.SuccessResponse{
			Success: true,
			Message: "Blog rejected, Aibys is improving the content based on your feedback",
			Data:    blog,
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Blog rejected successfully",
		Data:    blog,
	})
}

func regenerateBlog(blog *models.Blog, comment string) {
	log.Printf("[REGENERATE] blog id: %d, title: %s", blog.Id, blog.Title)

	prompt := fmt.Sprintf(`Kamu adalah Aibys, AI Assistant dari Arlchoose.

Kamu sebelumnya menulis artikel berjudul: "%s"

Deskripsi sebelumnya:
%s

Konten artikel sebelumnya:
%s

Artikel ini ditolak dengan catatan berikut dari editor:
"%s"

Tugasmu: Perbaiki artikel di atas sesuai catatan yang diberikan.

Instruksi:
- Perbaiki SESUAI catatan penolakan, jangan abaikan
- Tetap tulis dalam Bahasa Indonesia
- Format menggunakan HTML (h2, h3, p, ul, li, strong, em)
- JANGAN gunakan backtick atau markdown, HANYA HTML murni
- JANGAN ubah judul artikel
- Pertahankan fakta dan informasi yang sudah benar

Format response:
---DESCRIPTION---
[deskripsi singkat artikel yang sudah diperbaiki]
---CONTENT---
[konten artikel HTML yang sudah diperbaiki]`, blog.Title, blog.Description, blog.Content, comment)

	response, err := helpers.AskOllama(prompt)
	if err != nil {
		log.Printf("[REGENERATE ERROR] blog id: %d, err: %v", blog.Id, err)
		broadcastSSE(fmt.Sprintf(`{"type":"regenerate_done","blog_id":%d,"success":false}`, blog.Id))
		return
	}

	description, content := helpers.ParseOllamaResponse(response)
	if content == "" {
		log.Printf("[REGENERATE ERROR] empty content for blog id: %d", blog.Id)
		broadcastSSE(fmt.Sprintf(`{"type":"regenerate_done","blog_id":%d,"success":false}`, blog.Id))
		return
	}

	content = helpers.CleanAIOutput(content)
	description = helpers.CleanAIOutput(description)

	var freshBlog models.Blog
	if err := database.DB.First(&freshBlog, blog.Id).Error; err != nil {
		return
	}

	freshBlog.Description = description
	freshBlog.Content = content
	freshBlog.Status = "pending"
	freshBlog.RejectComment = ""

	if err := database.DB.Select("description", "content", "status", "reject_comment").Save(&freshBlog).Error; err != nil {
		log.Printf("[REGENERATE DB ERROR] blog id: %d, err: %v", blog.Id, err)
		broadcastSSE(fmt.Sprintf(`{"type":"regenerate_done","blog_id":%d,"success":false}`, blog.Id))
		return
	}

	log.Printf("[REGENERATE OK] blog id: %d done", blog.Id)
	broadcastSSE(fmt.Sprintf(`{"type":"regenerate_done","blog_id":%d,"success":true}`, blog.Id))
}
