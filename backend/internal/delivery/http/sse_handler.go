package http

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/pkg/sse"
)

// SSEHandler mengelola endpoint Server-Sent Events untuk realtime notification.
type SSEHandler struct {
	hub *sse.Hub
}

// NewSSEHandler membuat SSEHandler baru.
func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

// RegisterRoutes mendaftarkan route SSE.
func (h *SSEHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/sse/notifications", h.Stream)
}

// Stream adalah SSE endpoint yang di-subscribe oleh browser admin.
// Browser harus mengirim JWT token via query param ?token=... karena EventSource
// tidak mendukung custom header.
// Optional: ?employee_id=<uuid> untuk mengaktifkan personalized routing per employee.
func (h *SSEHandler) Stream(c *gin.Context) {
	// Ambil company_id dari context (sudah diset oleh middleware)
	companyID := ""
	if val, exists := c.Get("companyID"); exists {
		if cid, ok := val.(uuid.UUID); ok {
			companyID = cid.String()
		}
	}

	// Ambil user ID dari JWT context
	userID := ""
	if val, exists := c.Get("userID"); exists {
		if uid, ok := val.(uuid.UUID); ok {
			userID = uid.String()
		}
	}

	// Ambil role dari JWT context
	role := ""
	if val, exists := c.Get("role"); exists {
		if r, ok := val.(string); ok {
			role = r
		}
	}

	// employee_id dikirim frontend sebagai query param (auth store sudah menyimpannya)
	employeeID := c.Query("employee_id")

	// branchID dibaca dari JWT context (middleware sudah set dari employee.BranchID)
	// fallback ke query param jika context kosong
	branchID := ""
	if val, exists := c.Get("branchID"); exists {
		if bid, ok := val.(uuid.UUID); ok {
			branchID = bid.String()
		}
	}
	if branchID == "" {
		branchID = c.Query("branch_id") // Fallback dari frontend
	}

	clientID := fmt.Sprintf("%s-%d", companyID, time.Now().UnixNano())
	client := &sse.Client{
		ID:         clientID,
		CompanyID:  companyID,
		UserID:     userID,
		EmployeeID: employeeID,
		Role:       role,
		BranchID:   branchID,
		Send:       make(chan []byte, 16),
	}

	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Penting untuk Nginx reverse proxy

	// Kirim event "connected" sebagai handshake awal
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"client_id\":\"%s\"}\n\n", clientID)
	c.Writer.Flush()

	// Ticker untuk heartbeat agar koneksi tidak timeout
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Listen untuk disconnect dari client
	notify := c.Request.Context().Done()

	for {
		select {
		case <-notify:
			// Client disconnect
			return
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			// Deteksi tipe event dari JSON payload untuk event name SSE yang dinamis
			var raw struct {
				Type string `json:"type"`
			}
			eventType := "nota_fisik" // default fallback
			_ = json.Unmarshal(msg, &raw)
			if raw.Type != "" {
				eventType = raw.Type
			}
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", eventType, string(msg))
			c.Writer.Flush()
		case <-ticker.C:
			// Heartbeat untuk menjaga koneksi tetap hidup
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()
		}
	}
}
