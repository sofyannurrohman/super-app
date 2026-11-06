package sse

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

// NotaFisikEvent adalah payload event yang dipublish ke Redis dan dikirim ke client SSE.
type NotaFisikEvent struct {
	TransactionID   string   `json:"transaction_id"`
	ReceiptNo       string   `json:"receipt_no"`
	Salesman        string   `json:"salesman"`
	StoreName       string   `json:"store_name"`
	ImageURL        string   `json:"image_url"`
	CompanyID       string   `json:"company_id"`
	CreatedAt       string   `json:"created_at"`
	TargetBranchIDs []string `json:"target_branch_ids,omitempty"`
	TargetRoles     []string `json:"target_roles,omitempty"`
}

const (
	// RedisChannel adalah nama channel Redis Pub/Sub untuk event nota fisik.
	RedisChannel = "nota_fisik_uploaded"
	// RealtimeChannel adalah nama channel Redis Pub/Sub untuk event realtime umum.
	RealtimeChannel = "realtime_notifications"
	// SalesmanLocationChannel adalah nama channel Redis Pub/Sub untuk update lokasi salesman.
	SalesmanLocationChannel = "salesman_location_updates"
)

// NotificationEvent adalah payload event umum untuk realtime notification.
// Field targeting bersifat opsional — jika semua kosong → broadcast ke seluruh company (backward-compatible).
// Logika AND: jika TargetBranchIDs ada, client HARUS berada di branch tersebut,
// kemudian dicek kondisi role/employee/user (OR logic).
type NotificationEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "nota_fisik", "leave", "overtime", "warehouse_new_sj", dll.
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	CompanyID string                 `json:"company_id"`
	CreatedAt string                 `json:"created_at"`
	Data      map[string]interface{} `json:"data,omitempty"`

	// Personalized targeting — kosong berarti broadcast ke semua dalam company.
	TargetEmployeeIDs []string `json:"target_employee_ids,omitempty"`
	TargetUserIDs     []string `json:"target_user_ids,omitempty"`
	TargetRoles       []string `json:"target_roles,omitempty"`
	// Branch targeting (AND logic dengan TargetRoles) — kosong berarti tidak filter per branch.
	TargetBranchIDs []string `json:"target_branch_ids,omitempty"`
}

// Client merepresentasikan satu koneksi SSE dari browser admin.
type Client struct {
	ID         string
	CompanyID  string
	Send       chan []byte
	UserID     string // UUID user yang sedang login
	EmployeeID string // UUID employee terkait (dari query param saat connect)
	Role       string // Role user untuk role-based routing
	BranchID   string // UUID branch user (dari JWT context, di-set middleware)
}

// Hub mengelola semua client SSE dan subscription Redis.
type Hub struct {
	mu          sync.RWMutex
	clients     map[string]*Client // key: client ID
	redisClient *redis.Client
}

// NewHub membuat Hub baru dan memulai Redis subscriber.
func NewHub(rdb *redis.Client) *Hub {
	h := &Hub{
		clients:     make(map[string]*Client),
		redisClient: rdb,
	}
	go h.subscribeRedis()
	return h
}

// Register menambahkan client baru ke hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.ID] = c
	log.Printf("[SSE] Client registered: %s (company: %s, branch: %s, role: %s)", c.ID, c.CompanyID, c.BranchID, c.Role)
}

// Unregister menghapus client dari hub dan menutup channel-nya.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c.ID]; ok {
		delete(h.clients, c.ID)
		close(c.Send)
		log.Printf("[SSE] Client unregistered: %s", c.ID)
	}
}

// PublishNotaFisik mempublish event ke Redis Pub/Sub.
func (h *Hub) PublishNotaFisik(ctx context.Context, event NotaFisikEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return h.redisClient.Publish(ctx, RedisChannel, payload).Err()
}

// PublishNotification mempublish event realtime umum ke Redis Pub/Sub.
// Mendukung personalized routing via TargetBranchIDs, TargetRoles, TargetEmployeeIDs, TargetUserIDs.
func (h *Hub) PublishNotification(ctx context.Context, event NotificationEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return h.redisClient.Publish(ctx, RealtimeChannel, payload).Err()
}

// matchesTarget menentukan apakah client harus menerima event berdasarkan targeting fields.
//
// Logika:
//  1. Jika tidak ada target sama sekali → broadcast (backward-compatible).
//  2. Jika TargetBranchIDs ada → client HARUS match salah satu branch (filter AND).
//     Setelah branch lulus, cek kondisi lain (role/employee/user) dengan OR logic.
//  3. Jika tidak ada TargetBranchIDs → langsung cek kondisi lain (OR logic).
func matchesTarget(client *Client, targetEmployeeIDs, targetUserIDs, targetRoles, targetBranchIDs []string) bool {
	hasBranch := len(targetBranchIDs) > 0
	hasEmployee := len(targetEmployeeIDs) > 0
	hasUser := len(targetUserIDs) > 0
	hasRole := len(targetRoles) > 0

	// Tidak ada target sama sekali → broadcast ke semua
	if !hasBranch && !hasEmployee && !hasUser && !hasRole {
		return true
	}

	// Filter branch (AND): jika ada TargetBranchIDs, client HARUS berada di salah satunya
	// Superadmin atau Cashier Pusat tanpa branch binding khusus dapat menerima secara global
	if hasBranch {
		branchMatch := false
		if client.BranchID == "" && (client.Role == "superadmin" || client.Role == "cashier_pusat") {
			branchMatch = true
		} else {
			for _, bid := range targetBranchIDs {
				if bid == client.BranchID {
					branchMatch = true
					break
				}
			}
		}
		if !branchMatch {
			return false // Branch tidak match → tolak
		}
		// Branch lulus → jika tidak ada kondisi lain, langsung kirim
		if !hasEmployee && !hasUser && !hasRole {
			return true
		}
	}

	// Filter kondisi lain (OR logic): employee / user / role
	if hasEmployee && client.EmployeeID != "" {
		for _, eid := range targetEmployeeIDs {
			if eid == client.EmployeeID {
				return true
			}
		}
	}
	if hasUser && client.UserID != "" {
		for _, uid := range targetUserIDs {
			if uid == client.UserID {
				return true
			}
		}
	}
	if hasRole && client.Role != "" {
		for _, r := range targetRoles {
			if r == client.Role {
				return true
			}
		}
	}
	return false
}

// subscribeRedis listen ke Redis channel dan broadcast ke semua client SSE yang relevan.
func (h *Hub) subscribeRedis() {
	ctx := context.Background()
	sub := h.redisClient.Subscribe(ctx, RedisChannel, RealtimeChannel, SalesmanLocationChannel)
	defer sub.Close()

	log.Printf("[SSE] Subscribed to Redis channels: %s, %s, %s", RedisChannel, RealtimeChannel, SalesmanLocationChannel)

	ch := sub.Channel()
	for msg := range ch {
		// Parse CompanyID dan semua optional targeting fields dari payload
		var fullEvent struct {
			CompanyID         string   `json:"company_id"`
			TargetEmployeeIDs []string `json:"target_employee_ids"`
			TargetUserIDs     []string `json:"target_user_ids"`
			TargetRoles       []string `json:"target_roles"`
			TargetBranchIDs   []string `json:"target_branch_ids"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &fullEvent); err != nil {
			log.Printf("[SSE] Failed to parse event payload: %v", err)
			continue
		}

		payload := []byte(msg.Payload)
		h.mu.RLock()
		for _, client := range h.clients {
			// Filter per company
			if client.CompanyID != "" && client.CompanyID != fullEvent.CompanyID {
				continue
			}
			// Filter personalized (branch AND role / employee / user)
			if !matchesTarget(client,
				fullEvent.TargetEmployeeIDs,
				fullEvent.TargetUserIDs,
				fullEvent.TargetRoles,
				fullEvent.TargetBranchIDs,
			) {
				continue
			}
			select {
			case client.Send <- payload:
			default:
				// Buffer penuh, lewati (client lambat)
				log.Printf("[SSE] Client %s buffer full, skipping", client.ID)
			}
		}
		h.mu.RUnlock()
	}
}
