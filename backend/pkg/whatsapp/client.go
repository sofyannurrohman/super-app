package whatsapp

import (
	"context"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// Client adalah wrapper untuk whatsmeow yang mengelola koneksi WA.
type Client struct {
	wac *whatsmeow.Client
}

// NewClient membuat WA client baru.
//   - Sesi disimpan di SQLite (WA_SESSION_DB, default: ./wa_session.db)
//   - Jika belum pernah login, akan menampilkan QR code di terminal
//   - Jika sudah login sebelumnya, akan reconnect otomatis
func NewClient(ctx context.Context) (*Client, error) {
	dbPath := os.Getenv("WA_SESSION_DB")
	if dbPath == "" {
		dbPath = "./wa_session.db"
	}

	dbLog := waLog.Stdout("WA-DB", "WARN", true)
	container, err := sqlstore.New(ctx, "sqlite3",
		fmt.Sprintf("file:%s?_foreign_keys=on", dbPath), dbLog)
	if err != nil {
		return nil, fmt.Errorf("sqlstore init: %w", err)
	}

	// Ambil device store pertama (atau buat baru jika belum ada)
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("get device store: %w", err)
	}

	clientLog := waLog.Stdout("WA-Client", "WARN", true)
	wac := whatsmeow.NewClient(deviceStore, clientLog)

	if wac.Store.ID == nil {
		// Belum pernah login — tampilkan QR code di terminal
		qrChan, _ := wac.GetQRChannel(ctx)
		if err := wac.Connect(); err != nil {
			return nil, fmt.Errorf("wa connect (new): %w", err)
		}

		fmt.Println("\n" + repeatChar("═", 60))
		fmt.Println("  🔲 WHATSAPP LOGIN — Scan QR Code di bawah ini")
		fmt.Println("  📱 Buka WA → ⋮ → Linked Devices → Link a Device")
		fmt.Println(repeatChar("═", 60))

		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else if evt.Event == "success" {
				fmt.Println("\n✅ WhatsApp berhasil terhubung dan sesi disimpan!")
				fmt.Println(repeatChar("═", 60) + "\n")
				break
			} else if evt.Event == "timeout" {
				return nil, fmt.Errorf("QR code timeout — jalankan ulang server untuk scan baru")
			} else {
				fmt.Printf("   WA event: %s\n", evt.Event)
			}
		}
	} else {
		// Sudah pernah login — reconnect dari SQLite
		if err := wac.Connect(); err != nil {
			return nil, fmt.Errorf("wa reconnect: %w", err)
		}
		fmt.Printf("✅ WhatsApp reconnected (JID: %s)\n", wac.Store.ID.String())
	}

	return &Client{wac: wac}, nil
}

// SendToGroup mengirim pesan teks ke group WhatsApp.
// groupJID format: "628xxx-1234567890@g.us"
func (c *Client) SendToGroup(groupJID, message string) error {
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("parse JID '%s': %w", groupJID, err)
	}

	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	_, err = c.wac.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return fmt.Errorf("send message ke %s: %w", groupJID, err)
	}
	return nil
}

// GetGroups mengembalikan daftar semua group WA yang diikuti.
// Berguna untuk menemukan Group JID target.
func (c *Client) GetGroups(ctx context.Context) ([]GroupInfo, error) {
	groups, err := c.wac.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("get joined groups: %w", err)
	}

	var result []GroupInfo
	for _, g := range groups {
		result = append(result, GroupInfo{
			JID:  g.JID.String(),
			Name: g.Name,
		})
	}
	return result, nil
}

// Disconnect menutup koneksi WA dengan bersih (panggil saat shutdown).
func (c *Client) Disconnect() {
	if c.wac != nil {
		c.wac.Disconnect()
	}
}

// GroupInfo berisi informasi dasar group WhatsApp.
type GroupInfo struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
}

func repeatChar(ch string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += ch
	}
	return result
}
