package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TripRegionType menentukan skema ongkosan harian driver
type TripRegionType string

const (
	TripRegionInProvince  TripRegionType = "IN_PROVINCE"  // Dalam Provinsi: Rp 100.000 / hari
	TripRegionOutProvince TripRegionType = "OUT_PROVINCE" // Luar Provinsi (Sumatera, dll): Rp 150.000 / hari
)

// TripStatus status siklus perjalanan pengiriman
type TripStatus string

const (
	TripStatusDraft     TripStatus = "DRAFT"
	TripStatusScheduled TripStatus = "SCHEDULED"
	TripStatusInTransit TripStatus = "IN_TRANSIT"
	TripStatusArrived   TripStatus = "ARRIVED_DESTINATION"
	TripStatusReturning TripStatus = "RETURNING"
	TripStatusCompleted TripStatus = "COMPLETED"
	TripStatusCancelled TripStatus = "CANCELLED"
)

// TransportTrip merepresentasikan data penugasan dan riwayat perjalanan ekspedisi armada
type TransportTrip struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TripNumber string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"trip_number"`
	CompanyID  uuid.UUID `gorm:"type:uuid;not null" json:"company_id"`
	BranchID   *uuid.UUID `gorm:"type:uuid" json:"branch_id"`
	VehicleID  uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id"`
	DriverID   uuid.UUID `gorm:"type:uuid;not null" json:"driver_id"`
	HelperID   *uuid.UUID `gorm:"type:uuid" json:"helper_id"`

	// Rute & Wilayah
	OriginCity          string         `gorm:"type:varchar(100);not null" json:"origin_city"`
	OriginProvince      string         `gorm:"type:varchar(100);not null" json:"origin_province"`
	DestinationCity     string         `gorm:"type:varchar(100);not null" json:"destination_city"`
	DestinationProvince string         `gorm:"type:varchar(100);not null" json:"destination_province"`
	RegionType          TripRegionType `gorm:"type:varchar(20);not null;default:'IN_PROVINCE'" json:"region_type"`

	// Durasi & Ongkos Harian Driver
	DepartureDate       time.Time  `gorm:"type:date;not null" json:"departure_date"`
	EstimatedReturnDate time.Time  `gorm:"type:date;not null" json:"estimated_return_date"`
	ActualReturnDate    *time.Time `gorm:"type:date" json:"actual_return_date"`
	EstimatedDays       int        `gorm:"type:int;not null;default:1" json:"estimated_days"`
	ActualDays          *int       `gorm:"type:int" json:"actual_days"`
	DailyAllowanceRate  float64    `gorm:"type:decimal(15,2);not null" json:"daily_allowance_rate"` // 100.000 (Dalam Provinsi) atau 150.000 (Luar Provinsi)
	TotalDriverFee      float64    `gorm:"type:decimal(15,2);not null" json:"total_driver_fee"`     // EstimatedDays * DailyAllowanceRate
	TotalHelperFee      float64    `gorm:"type:decimal(15,2);default:0" json:"total_helper_fee"`

	// Odometer KM Tracking
	StartOdometer int  `gorm:"type:int;default:0" json:"start_odometer"`
	EndOdometer   *int `gorm:"type:int" json:"end_odometer"`

	CargoDescription string     `gorm:"type:text" json:"cargo_description"`
	Status           TripStatus `gorm:"type:varchar(30);default:'SCHEDULED'" json:"status"`
	Notes            string     `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Vehicle       *Vehicle                  `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Driver        *Employee                 `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Helper        *Employee                 `gorm:"foreignKey:HelperID" json:"helper,omitempty"`
	Branch        *Branch                   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Backhauls     []TransportBackhaul       `gorm:"foreignKey:TripID" json:"backhauls,omitempty"`
	ExpenseClaims []TransportExpenseClaim   `gorm:"foreignKey:TripID" json:"expense_claims,omitempty"`
	Inspection    *TransportFleetInspection `gorm:"foreignKey:TripID" json:"inspection,omitempty"`
}

func (t *TransportTrip) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

// ApprovalStatus status approval persetujuan
type TransportApprovalStatus string

const (
	TransportApprovalPending  TransportApprovalStatus = "PENDING"
	TransportApprovalApproved TransportApprovalStatus = "APPROVED"
	TransportApprovalRejected TransportApprovalStatus = "REJECTED"
)

// TransportBackhaul merepresentasikan muatan kembali / pengiriman titipan dari lokasi A ke B
type TransportBackhaul struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TripID              *uuid.UUID `gorm:"type:uuid" json:"trip_id"`
	BackhaulNumber      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"backhaul_number"`
	DriverID            uuid.UUID `gorm:"type:uuid;not null" json:"driver_id"`
	VehicleID           uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id"`
	OriginLocation      string    `gorm:"type:varchar(255);not null" json:"origin_location"`
	DestinationLocation string    `gorm:"type:varchar(255);not null" json:"destination_location"`
	CustomerName        string    `gorm:"type:varchar(150);not null" json:"customer_name"`
	CustomerPhone       string    `gorm:"type:varchar(50)" json:"customer_phone"`
	CargoName           string    `gorm:"type:varchar(255);not null" json:"cargo_name"`
	CargoWeightKg       float64   `gorm:"type:decimal(10,2);default:0" json:"cargo_weight_kg"`

	// Perhitungan Biaya & Fee 10%
	FreightCost   float64 `gorm:"type:decimal(15,2);not null" json:"freight_cost"`      // Total biaya ongkos kirim muatan kembali
	FeePercentage float64 `gorm:"type:decimal(5,2);default:10.0" json:"fee_percentage"` // 10%
	FeeAmount     float64 `gorm:"type:decimal(15,2);not null" json:"fee_amount"`        // 10% * FreightCost
	NetAmount     float64 `gorm:"type:decimal(15,2);not null" json:"net_amount"`        // FreightCost - FeeAmount

	AttachmentProofURL string                  `gorm:"type:text" json:"attachment_proof_url"` // Foto surat jalan / nota muatan
	Status             TransportApprovalStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ApprovedBy         *uuid.UUID              `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt         *time.Time              `json:"approved_at"`
	RejectionReason    string                  `gorm:"type:text" json:"rejection_reason"`
	Notes              string                  `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Trip     *TransportTrip `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	Driver   *Employee      `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Vehicle  *Vehicle       `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Approver *Employee      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (b *TransportBackhaul) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

// ExpenseCategory kategori klaim biaya perjalanan
type ExpenseCategory string

const (
	ExpenseCategoryBBM       ExpenseCategory = "BBM"
	ExpenseCategorySparepart ExpenseCategory = "SPAREPART_ARMADA"
	ExpenseCategoryToll      ExpenseCategory = "TOLL_PARKING"
	ExpenseCategoryEmergency ExpenseCategory = "EMERGENCY_REPAIR"
	ExpenseCategoryOther     ExpenseCategory = "OTHER"
)

// TransportExpenseClaim klaim struk nota BBM dan sparepart armada dengan approval
type TransportExpenseClaim struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TripID          *uuid.UUID      `gorm:"type:uuid" json:"trip_id"`
	VehicleID       uuid.UUID       `gorm:"type:uuid;not null" json:"vehicle_id"`
	DriverID        uuid.UUID       `gorm:"type:uuid;not null" json:"driver_id"`
	ClaimNumber     string          `gorm:"type:varchar(50);uniqueIndex;not null" json:"claim_number"`
	Category        ExpenseCategory `gorm:"type:varchar(30);not null" json:"category"`
	MerchantName    string          `gorm:"type:varchar(150)" json:"merchant_name"` // SPBU / Toko Sparepart / Bengkel
	FuelLiter       *float64        `gorm:"type:decimal(10,2)" json:"fuel_liter"`
	CurrentOdometer int             `gorm:"type:int;default:0" json:"current_odometer"`
	Amount          float64         `gorm:"type:decimal(15,2);not null" json:"amount"`
	ReceiptImageURL string          `gorm:"type:text;not null" json:"receipt_image_url"` // Foto struk nota fisik
	Description     string          `gorm:"type:text" json:"description"`

	// Approval & Disbursement
	Status          TransportApprovalStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ApprovedBy      *uuid.UUID              `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt      *time.Time              `json:"approved_at"`
	RejectionReason string                  `gorm:"type:text" json:"rejection_reason"`

	DisbursementStatus string     `gorm:"type:varchar(20);default:'UNPAID'" json:"disbursement_status"` // UNPAID, PAID
	DisbursedAt        *time.Time `json:"disbursed_at"`
	DisbursedBy        *uuid.UUID `gorm:"type:uuid" json:"disbursed_by"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Trip     *TransportTrip `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	Vehicle  *Vehicle       `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Driver   *Employee      `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Approver *Employee      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (c *TransportExpenseClaim) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

// FollowUpAction tindakan pasca inspeksi armada
type FollowUpAction string

const (
	FollowUpNone            FollowUpAction = "NO_ACTION_REQUIRED"     // Kondisi Normal Siap Jalan
	FollowUpMinorRepair     FollowUpAction = "NEEDS_MINOR_REPAIR"     // Perlu Servis Ringan
	FollowUpPartReplacement FollowUpAction = "NEEDS_PART_REPLACEMENT" // Perlu Ganti Sparepart
	FollowUpMajorOverhaul   FollowUpAction = "NEEDS_OVERHAUL"         // Perlu Perbaikan Besar
)

type MaintenanceFollowUpStatus string

const (
	FollowUpStatusPending    MaintenanceFollowUpStatus = "PENDING_REPAIR"
	FollowUpStatusInProgress MaintenanceFollowUpStatus = "IN_PROGRESS"
	FollowUpStatusCompleted  MaintenanceFollowUpStatus = "COMPLETED"
)

// TransportFleetInspection laporan kondisi armada pasca perjalanan dan tindak lanjut perbaikan/pergantian part
type TransportFleetInspection struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TripID            *uuid.UUID `gorm:"type:uuid" json:"trip_id"`
	VehicleID         uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id"`
	InspectorDriverID uuid.UUID `gorm:"type:uuid;not null" json:"inspector_driver_id"`
	InspectionDate    time.Time `gorm:"type:date;not null" json:"inspection_date"`
	ReturnOdometer    int       `gorm:"type:int;not null" json:"return_odometer"`

	// Checklist Kondisi (GOOD, FAIR, DAMAGED, CRITICAL)
	BrakeCondition   string `gorm:"type:varchar(20);default:'GOOD'" json:"brake_condition"`
	TireCondition    string `gorm:"type:varchar(20);default:'GOOD'" json:"tire_condition"`
	EngineCondition  string `gorm:"type:varchar(20);default:'GOOD'" json:"engine_condition"`
	OilCondition     string `gorm:"type:varchar(20);default:'GOOD'" json:"oil_condition"`
	BatteryCondition string `gorm:"type:varchar(20);default:'GOOD'" json:"battery_condition"`
	ElectricalLights string `gorm:"type:varchar(20);default:'GOOD'" json:"electrical_lights"`
	BodyCondition    string `gorm:"type:varchar(20);default:'GOOD'" json:"body_condition"`

	DriverRemarks      string `gorm:"type:text" json:"driver_remarks"`
	PhotoConditionURLs string `gorm:"type:text" json:"photo_condition_urls"` // Path URL gambar dipisahkan koma atau JSON

	// Tindak Lanjut
	FollowUpAction FollowUpAction            `gorm:"type:varchar(30);default:'NO_ACTION_REQUIRED'" json:"follow_up_action"`
	PartsToReplace string                    `gorm:"type:text" json:"parts_to_replace"` // Nama part yang perlu diganti
	WorkshopName   string                    `gorm:"type:varchar(150)" json:"workshop_name"`
	EstimatedCost  float64                   `gorm:"type:decimal(15,2);default:0" json:"estimated_cost"`
	ActualCost     float64                   `gorm:"type:decimal(15,2);default:0" json:"actual_cost"`
	FollowUpStatus MaintenanceFollowUpStatus `gorm:"type:varchar(30);default:'PENDING_REPAIR'" json:"follow_up_status"`
	RepairedAt     *time.Time                `json:"repaired_at"`
	RepairNotes    string                    `gorm:"type:text" json:"repair_notes"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Trip    *TransportTrip `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	Vehicle *Vehicle       `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Driver  *Employee      `gorm:"foreignKey:InspectorDriverID" json:"driver,omitempty"`
}

func (i *TransportFleetInspection) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return
}

// TransportDriverAttendance log absensi & tugas driver luar kota
type TransportDriverAttendance struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TripID         uuid.UUID `gorm:"type:uuid;not null" json:"trip_id"`
	DriverID       uuid.UUID `gorm:"type:uuid;not null" json:"driver_id"`
	AttendanceDate time.Time `gorm:"type:date;not null" json:"attendance_date"`
	Status         string    `gorm:"type:varchar(20);default:'ON_DUTY_TRIP'" json:"status"` // ON_DUTY_TRIP, REST, STANDBY
	CurrentCity    string    `gorm:"type:varchar(100)" json:"current_city"`
	CheckinLat     float64   `gorm:"type:decimal(10,8)" json:"checkin_lat"`
	CheckinLong    float64   `gorm:"type:decimal(11,8)" json:"checkin_long"`
	PhotoURL       string    `gorm:"type:text" json:"photo_url"`
	DailyAllowance float64   `gorm:"type:decimal(15,2);not null" json:"daily_allowance"` // 100k / 150k
	Notes          string    `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	Trip   *TransportTrip `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	Driver *Employee      `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
}

func (a *TransportDriverAttendance) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
