package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Store struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id" form:"id"`
	CompanyID           uuid.UUID  `gorm:"type:uuid;not null" json:"company_id" form:"company_id"`
	CreatedByEmployeeID *uuid.UUID `gorm:"type:uuid" json:"created_by_employee_id" form:"-"`
	AssignedEmployeeID  *uuid.UUID `gorm:"type:uuid" json:"assigned_employee_id" form:"assigned_employee_id"` // PENTING: Untuk Rayonisasi
	Name                string     `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	OwnerName           string     `gorm:"type:varchar(255)" json:"owner_name" form:"owner_name"`
	PhoneNumber         string     `gorm:"type:varchar(50)" json:"phone_number" form:"phone_number"`
	Address             string     `gorm:"type:text" json:"address" form:"address"`
	Latitude            float64    `gorm:"type:decimal(10,8)" json:"latitude" form:"latitude"`
	Longitude           float64    `gorm:"type:decimal(11,8)" json:"longitude" form:"longitude"`
	IsActive            bool       `gorm:"default:true" json:"is_active" form:"is_active"`
	Category            string     `gorm:"type:varchar(50);default:'TOKO_BARU'" json:"category" form:"category"`      // TOKO_BARU or TOKO_LAMA
	CustomerType        string     `gorm:"type:varchar(50);default:'TOKO'" json:"customer_type" form:"customer_type"` // WARUNG, TOKO, GROSIR, AGEN, DISTRIBUTOR, PASAR, MINIMARKET, SUPERMARKET, HOREKA, LAINNYA
	SubCustomerType     string     `gorm:"type:varchar(50)" json:"sub_customer_type" form:"sub_customer_type"`        // HOTEL, RESTORAN, CAFE, KATERING, RUMAH_MAKAN, LAINNYA
	DistributionCity    string     `gorm:"type:varchar(100)" json:"distribution_city" form:"distribution_city"`       // for DISTRIBUTOR
	PhotoURL            string     `gorm:"type:text" json:"photo_url" form:"photo_url"`
	Province            string     `gorm:"type:varchar(100)" json:"province" form:"province"`
	Regency             string     `gorm:"type:varchar(100)" json:"regency" form:"regency"`
	District            string     `gorm:"type:varchar(100)" json:"district" form:"district"`
	Village             string     `gorm:"type:varchar(100)" json:"village" form:"village"`

	// Fields untuk e-Faktur Pajak (Opsional)
	NPWP    *string `gorm:"type:varchar(20)" json:"npwp" form:"npwp"`
	NIK     *string `gorm:"type:varchar(20)" json:"nik" form:"nik"`
	TaxName *string `gorm:"type:varchar(255)" json:"tax_name" form:"tax_name"`

	FirstTransactionDate *time.Time `gorm:"type:date" json:"first_transaction_date" form:"first_transaction_date"`
	CreatedAt            time.Time  `gorm:"default:now()" json:"created_at" form:"-"`
	UpdatedAt            time.Time  `gorm:"default:now()" json:"updated_at" form:"-"`

	VisitDays      pq.Int64Array `gorm:"type:integer[]" json:"visit_days" form:"-"`
	VisitFrequency string        `gorm:"type:varchar(10)" json:"visit_frequency" form:"visit_frequency"` // F8, F4, F2
	// Sinkronisasi dari aplikasi katalog eksternal
	// NULL = toko asli Super App | Isi = toko yang diimport dari MySQL Branding via cronjob
	BrandingShopID *int `gorm:"-:migration;column:branding_shop_id;uniqueIndex;default:null" json:"branding_shop_id,omitempty" form:"-"`

	VerificationStatus string `gorm:"type:store_verification_status;default:'UNVERIFIED'" json:"verification_status" form:"verification_status"`

	// Fields for reporting (calculated via subqueries or aggregation)
	TotalReceivable    float64 `gorm:"-" json:"total_receivable"`
	OverdueCount       int     `gorm:"-" json:"overdue_count"`
	TransactionCount   int     `gorm:"-" json:"transaction_count"`
	LatestVerifierRole string  `gorm:"-" json:"latest_verifier_role"` // role of the last person who verified it

	// Real-time tracking fields for NOO & EC
	IsNooToday       bool    `gorm:"-" json:"is_noo_today"`
	HasOrderToday    bool    `gorm:"-" json:"has_order_today"`
	TodayOrderCount  int     `gorm:"-" json:"today_order_count"`
	TodayOrderAmount float64 `gorm:"-" json:"today_order_amount"`
	IsEffectiveCall  bool    `gorm:"-" json:"is_effective_call"`

	Company           *Company  `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty" form:"-"`
	AssignedEmployee  *Employee `gorm:"foreignKey:AssignedEmployeeID;constraint:OnDelete:SET NULL;" json:"assigned_employee,omitempty" form:"-"`
	CreatedByEmployee *Employee `gorm:"foreignKey:CreatedByEmployeeID;constraint:OnDelete:SET NULL;" json:"created_by_employee,omitempty" form:"-"`
	DistributorID     *uuid.UUID `gorm:"type:uuid;index" json:"distributor_id,omitempty" form:"distributor_id"`
	Distributor       *Store     `gorm:"foreignKey:DistributorID;constraint:OnDelete:SET NULL;" json:"distributor,omitempty" form:"-"`
}

func (s *Store) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

// DetermineCategory returns the store's category. If it is empty, it calculates it based on FirstTransactionDate or CreatedAt.
func (s *Store) DetermineCategory(transactionDate time.Time) string {
	if s.Category != "" {
		return s.Category
	}

	baseDate := s.CreatedAt
	if s.FirstTransactionDate != nil {
		baseDate = *s.FirstTransactionDate
	}

	diff := transactionDate.Sub(baseDate)
	if diff <= 90*24*time.Hour {
		return "TOKO_BARU"
	}
	return "TOKO_LAMA"
}

type CheckDuplicateStoreRequest struct {
	Name        string     `json:"name" form:"name"`
	PhoneNumber string     `json:"phone_number" form:"phone_number"`
	Latitude    float64    `json:"latitude" form:"latitude"`
	Longitude   float64    `json:"longitude" form:"longitude"`
	District    string     `json:"district" form:"district"`
	Regency     string     `json:"regency" form:"regency"`
	Province    string     `json:"province" form:"province"`
	ExcludeID   *uuid.UUID `json:"exclude_id" form:"exclude_id"`
}

type DuplicateStoreMatch struct {
	StoreID              uuid.UUID  `json:"store_id"`
	Name                 string     `json:"name"`
	PhoneNumber          string     `json:"phone_number"`
	Address              string     `json:"address"`
	DistanceMeters       float64    `json:"distance_meters"`
	NameSimilarity       float64    `json:"name_similarity"`
	TotalScore           float64    `json:"total_score"`
	MatchReasons         []string   `json:"match_reasons"`
	AssignedEmployeeID   *uuid.UUID `json:"assigned_employee_id,omitempty"`
	AssignedEmployeeName string     `json:"assigned_employee_name,omitempty"`
	AssignedBranchName   string     `json:"assigned_branch_name,omitempty"`
	IsAssignedToMe       bool       `json:"is_assigned_to_me"`
	PhotoURL             string     `json:"photo_url,omitempty"`
	IsActive             bool       `json:"is_active"`
}

type CheckDuplicateStoreResponse struct {
	HasPotentialDuplicate bool                  `json:"has_potential_duplicate"`
	Matches               []DuplicateStoreMatch `json:"matches"`
}

type StoreDistributionItem struct {
	ID                 uuid.UUID  `json:"id"`
	CompanyID          uuid.UUID  `json:"company_id"`
	Name               string     `json:"name"`
	OwnerName          string     `json:"owner_name"`
	PhoneNumber        string     `json:"phone_number"`
	Address            string     `json:"address"`
	Latitude           float64    `json:"latitude"`
	Longitude          float64    `json:"longitude"`
	IsActive           bool       `json:"is_active"`
	Category           string     `json:"category"`
	CustomerType       string     `json:"customer_type"`
	Province           string     `json:"province"`
	Regency            string     `json:"regency"`
	District           string     `json:"district"`
	Village            string     `json:"village"`
	AssignedEmployeeID *uuid.UUID `json:"assigned_employee_id"`
	EmployeeFirstName  string     `json:"employee_first_name"`
	EmployeeLastName   string     `json:"employee_last_name"`
	BranchID           *uuid.UUID `json:"branch_id"`
	BranchName         string     `json:"branch_name"`
	CreatedAt          time.Time  `json:"created_at"`
}

