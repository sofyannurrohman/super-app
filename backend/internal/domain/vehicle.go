package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vehicle struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id" form:"id"`
	CompanyID    uuid.UUID  `gorm:"type:uuid;not null" json:"company_id" form:"company_id"`
	BranchID     *uuid.UUID `gorm:"type:uuid" json:"branch_id" form:"branch_id"`
	Name         string     `gorm:"type:varchar(255);not null" json:"name" form:"name"`
	LicensePlate string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"license_plate" form:"license_plate"`
	Model        string     `gorm:"type:varchar(100)" json:"model" form:"model"`
	Type         string     `gorm:"type:varchar(50)" json:"type" form:"type"`                         // Truck, Car, Van, Motorcycle
	Status       string     `gorm:"type:varchar(20);default:'AVAILABLE'" json:"status" form:"status"` // AVAILABLE, IN_USE, MAINTENANCE, BROKEN
	Year         int        `gorm:"type:int" json:"year" form:"year"`
	Mileage      int        `gorm:"type:int;default:0" json:"mileage" form:"mileage"`
	Capacity     float64    `gorm:"type:decimal(10,2);default:0" json:"capacity" form:"capacity"`
	BedLength    float64    `gorm:"type:decimal(10,2);default:0" json:"bed_length" form:"bed_length"`
	BedWidth     float64    `gorm:"type:decimal(10,2);default:0" json:"bed_width" form:"bed_width"`
	BedHeight    float64    `gorm:"type:decimal(10,2);default:0" json:"bed_height" form:"bed_height"`
	BedVolume    float64    `gorm:"type:decimal(10,2);default:0" json:"bed_volume" form:"bed_volume"`
	ImageURL     string     `gorm:"type:varchar(255)" json:"image_url" form:"image_url"`
	PhotoFront   string     `gorm:"type:varchar(255)" json:"photo_front" form:"photo_front"`
	PhotoLeft    string     `gorm:"type:varchar(255)" json:"photo_left" form:"photo_left"`
	PhotoRight   string     `gorm:"type:varchar(255)" json:"photo_right" form:"photo_right"`
	PhotoBack    string     `gorm:"type:varchar(255)" json:"photo_back" form:"photo_back"`
	TaxExpiryDate *time.Time `gorm:"type:date" json:"tax_expiry_date" form:"tax_expiry_date"`
	GpsIMEI      string     `gorm:"type:varchar(50)" json:"gps_imei" form:"gps_imei"`
	CreatedAt    time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"default:now()" json:"updated_at"`

	Company          *Company                  `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Branch           *Branch                   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Components       []VehicleComponent        `gorm:"foreignKey:VehicleID" json:"components,omitempty"`
	ActiveAssignment *ActiveDeliveryAssignment `gorm:"-" json:"active_assignment,omitempty"`
}

type ActiveDeliveryType string

const (
	ActiveDeliveryTypeSalesBatch      ActiveDeliveryType = "SALES_DELIVERY"
	ActiveDeliveryTypeFactoryTransfer ActiveDeliveryType = "FACTORY_TRANSFER"
	ActiveDeliveryTypeMTDelivery      ActiveDeliveryType = "MT_DELIVERY"
	ActiveDeliveryTypeFactoryReturn   ActiveDeliveryType = "FACTORY_RETURN"
)

type ActiveDeliveryAssignment struct {
	DeliveryType    ActiveDeliveryType `json:"delivery_type"`
	TypeLabel       string             `json:"type_label"`
	ReferenceID     uuid.UUID          `json:"reference_id"`
	DONumber        string             `json:"do_number"`
	Status          string             `json:"status"`
	StatusLabel     string             `json:"status_label"`
	DriverName      string             `json:"driver_name"`
	DriverPhone     string             `json:"driver_phone,omitempty"`
	HelperName      string             `json:"helper_name,omitempty"`
	OriginName      string             `json:"origin_name"`
	DestinationName string             `json:"destination_name"`
	TotalItems      int                `json:"total_items"`
	StartedAt       *time.Time         `json:"started_at,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	RedirectPath    string             `json:"redirect_path"`
}

func (v *Vehicle) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return
}
