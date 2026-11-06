package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesReturnStatus string
type SalesReturnType string

const (
	ReturnStatusPending          SalesReturnStatus = "PENDING"
	ReturnStatusPendingWarehouse SalesReturnStatus = "PENDING_WAREHOUSE"
	ReturnStatusApproved         SalesReturnStatus = "APPROVED"
	ReturnStatusRejected         SalesReturnStatus = "REJECTED"
)

const (
	// POTONG_NOTA: Barang diretur, tagihan/nota dipotong senilai barang yang kembali.
	// Contoh: Customer menolak 2 kardus, faktur dikurangi senilai 2 kardus tersebut.
	ReturnTypePotongNota SalesReturnType = "POTONG_NOTA"

	// TUKAR_LANGSUNG: Barang rusak ditukar di tempat dengan barang sejenis yang bagus dari truk driver.
	// Tagihan tetap penuh. Barang rusak dibawa kembali ke gudang karantina.
	ReturnTypeTukarLangsung SalesReturnType = "TUKAR_LANGSUNG"

	// TUKAR_SUSULAN: Barang rusak tidak ada penggantinya saat ini, pengganti dikirim pada kunjungan berikutnya.
	// Tagihan tetap penuh. Membuat SalesSubsequentItem untuk pengiriman susulan.
	ReturnTypeTukarSusulan SalesReturnType = "TUKAR_SUSULAN"

	// TUKAR_PRODUK: Barang diretur dan diganti dengan PRODUK YANG BERBEDA (bukan produk yang sama).
	// Berguna untuk substitusi produk. ReplacementProductID wajib diisi.
	// Selisih harga diperhitungkan ke nota baru.
	ReturnTypeTukarProduk SalesReturnType = "TUKAR_PRODUK"

	// GARANSI: Klaim garansi produk rusak/cacat pabrik. Tidak memotong nota.
	// Produk pengganti dikirimkan oleh tim atau gudang secara terpisah.
	// ClaimReferenceNo wajib diisi. Status garansi dipantau di level retur ini.
	ReturnTypeGaransi SalesReturnType = "GARANSI"
)

// SalesReturn mencatat pengembalian barang dari pelanggan.
// Mendukung 5 tipe retur: POTONG_NOTA, TUKAR_LANGSUNG, TUKAR_SUSULAN, TUKAR_PRODUK, GARANSI.
type SalesReturn struct {
	ID            uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReturnNo      string            `gorm:"type:varchar(50);unique;not null" json:"return_no"`
	SalesOrderID  *uuid.UUID        `gorm:"type:uuid" json:"sales_order_id,omitempty"`
	TransactionID *uuid.UUID        `gorm:"type:uuid" json:"transaction_id"` // Link ke Invoice
	BranchID      uuid.UUID         `gorm:"type:uuid;not null" json:"branch_id"`
	EmployeeID    uuid.UUID         `gorm:"type:uuid;not null" json:"employee_id"` // Yang menginput (salesman/driver)
	ReturnType    string            `gorm:"type:varchar(30);default:'POTONG_NOTA'" json:"return_type"`
	Status        SalesReturnStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	TotalAmount   float64           `gorm:"type:decimal(15,2);default:0" json:"total_amount"` // Nilai yang memotong tagihan
	Notes         string            `gorm:"type:text" json:"notes"`

	// Khusus TUKAR_PRODUK: selisih harga yang harus ditambahkan/dikurangi ke faktur pengganti
	PriceDifference float64 `gorm:"type:decimal(15,2);default:0" json:"price_difference"`
	// Waktu penyesuaian faktur; stok fisik tetap menunggu persetujuan gudang.
	InvoiceAdjustedAt *time.Time `json:"invoice_adjusted_at,omitempty"`

	// Khusus GARANSI: nomor referensi klaim garansi & tanggal penyelesaian
	ClaimReferenceNo   string     `gorm:"type:varchar(100)" json:"claim_reference_no"`
	WarrantyResolvedAt *time.Time `json:"warranty_resolved_at,omitempty"`

	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	ApprovedByID *uuid.UUID `gorm:"type:uuid" json:"approved_by_id"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	SalesOrder  *SalesOrder       `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`
	Transaction *SalesTransaction `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
	Employee    *Employee         `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	ApprovedBy  *Employee         `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	Items       []SalesReturnItem `gorm:"foreignKey:SalesReturnID" json:"items,omitempty"`
}

// SalesReturnItem mencatat produk yang diretur.
type SalesReturnItem struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesReturnID uuid.UUID  `gorm:"type:uuid;not null" json:"sales_return_id"`
	ProductID     uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	SalesItemID   *uuid.UUID `gorm:"type:uuid" json:"sales_item_id,omitempty"`
	Quantity      int        `gorm:"type:int;not null" json:"quantity"`
	Price         float64    `gorm:"type:decimal(15,2);default:0" json:"price"`
	Reason        string     `gorm:"type:varchar(255)" json:"reason"`
	Condition     string     `gorm:"type:varchar(20);default:'DAMAGED'" json:"condition"` // 'GOOD', 'DAMAGED'

	// Khusus TUKAR_PRODUK: produk pengganti yang berbeda dari produk asal
	ReplacementProductID   *uuid.UUID `gorm:"type:uuid" json:"replacement_product_id"`
	ReplacementSalesItemID *uuid.UUID `gorm:"type:uuid" json:"replacement_sales_item_id,omitempty"`
	ReplacementQuantity    int        `gorm:"type:int;default:0" json:"replacement_quantity"`
	ReplacementPrice       float64    `gorm:"type:decimal(15,2);default:0" json:"replacement_price"`

	// Khusus GARANSI: foto bukti kerusakan sebagai dokumentasi klaim
	DamageEvidenceURL string `gorm:"type:text" json:"damage_evidence_url"`

	Product            *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ReplacementProduct *Product `gorm:"foreignKey:ReplacementProductID" json:"replacement_product,omitempty"`
}

func (sr *SalesReturn) BeforeCreate(tx *gorm.DB) (err error) {
	if sr.ID == uuid.Nil {
		sr.ID = uuid.New()
	}
	return
}
