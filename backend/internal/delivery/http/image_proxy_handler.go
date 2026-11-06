package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ImageProxyHandler struct {
	httpClient *http.Client
}

func NewImageProxyHandler() *ImageProxyHandler {
	return &ImageProxyHandler{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (h *ImageProxyHandler) SetupPublicRoutes(r *gin.RouterGroup) {
	r.GET("/image-proxy", h.ProxyImage)
	r.OPTIONS("/image-proxy", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Status(http.StatusNoContent)
	})
}

// ProxyImage serves cross-origin images to Flutter Web or other clients
// without CORS restrictions.
// Usage: GET /api/v1/image-proxy?url=https%3A%2F%2Fexample.com%2Fstorage%2Ffoto-toko%2Fxxx.webp
func (h *ImageProxyHandler) ProxyImage(c *gin.Context) {
	rawURL := strings.TrimSpace(c.Query("url"))
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameter 'url' diperlukan"})
		return
	}

	// Basic validation of URL structure
	parsedURL, err := url.Parse(rawURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL tidak valid"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Gagal membuat request: %v", err)})
		return
	}

	// Mimic authentic browser headers so Hostinger CDN & Cloudflare allow image retrieval
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Referer", fmt.Sprintf("%s://%s/", parsedURL.Scheme, parsedURL.Host))
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="125", "Chromium";v="125", "Not.A/Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "image")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Gagal mengambil gambar: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("Remote server mengembalikan status %d", resp.StatusCode)})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "image/") {
		// Fallback content type detection based on extension
		lowerPath := strings.ToLower(parsedURL.Path)
		if strings.HasSuffix(lowerPath, ".webp") {
			contentType = "image/webp"
		} else if strings.HasSuffix(lowerPath, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(lowerPath, ".jpg") || strings.HasSuffix(lowerPath, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(lowerPath, ".svg") {
			contentType = "image/svg+xml"
		} else {
			contentType = "image/jpeg"
		}
	}

	// Set CORS & Cache headers
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Cache-Control", "public, max-age=86400, immutable")
	c.Header("Content-Type", contentType)

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Header("Content-Length", contentLength)
	}

	// Stream response
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, resp.Body)
}
