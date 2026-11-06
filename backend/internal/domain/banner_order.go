package domain

import (
	"time"

	"github.com/google/uuid"
)

type BannerOrder struct {
	ID                    uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CompanyID             uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"`
	EmployeeID            uuid.UUID  `gorm:"type:uuid;not null" json:"employee_id"`
	StoreID               *uuid.UUID `gorm:"type:uuid" json:"store_id"`
	CampaignName          string     `gorm:"type:varchar(255)" json:"campaign_name"`
	StoreName             string     `gorm:"type:varchar(255)" json:"store_name"`
	Location              string     `gorm:"type:varchar(255)" json:"location"`
	BannerType            string     `gorm:"type:varchar(100)" json:"banner_type"`
	Size                  float64    `gorm:"type:decimal(10,2);not null" json:"size"`
	DesignURL             string     `gorm:"type:text" json:"design_url"`
	DesignLink            *string    `gorm:"column:design_link" json:"design_link"`
	DesignFileName        *string    `gorm:"column:design_file_name;type:varchar(500)" json:"design_file_name"`
	DocumentationImageURL string     `gorm:"type:text" json:"documentation_image_url"`
	Notes                 *string    `gorm:"type:text" json:"notes"`
	Status                string     `gorm:"type:varchar(50);default:'PENDING'" json:"status"`
	EditorID              *uuid.UUID `gorm:"type:uuid" json:"editor_id"`
	VendorID              *uuid.UUID `gorm:"type:uuid" json:"vendor_id"`
	OperatorID            *uuid.UUID `gorm:"type:uuid" json:"operator_id"`
	DesignerID            *uuid.UUID `gorm:"type:uuid" json:"designer_id"`
	InstallerID           *uuid.UUID `gorm:"type:uuid" json:"installer_id"`
	CreatedAt             time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	MySQLUpdatedAt     *time.Time `gorm:"column:mysql_updated_at" json:"mysql_updated_at"`
	MySQLBannerID      *int64     `gorm:"column:mysql_banner_id;index" json:"mysql_banner_id"`
	MySQLSalesName     *string    `gorm:"column:mysql_sales_name" json:"mysql_sales_name"`
	MySQLEditorName    *string    `gorm:"column:mysql_editor_name" json:"mysql_editor_name"`
	Qty                int        `gorm:"column:qty;default:1" json:"qty"`
	SizeLength         float64    `gorm:"column:size_length;default:0" json:"size_length"`
	SizeWidth          float64    `gorm:"column:size_width;default:0" json:"size_width"`
	IsBihun            bool       `gorm:"column:is_bihun;default:false" json:"is_bihun"`
	RevisionNote       *string    `gorm:"column:revision_note" json:"revision_note"`
	IsDesignApproved   bool       `gorm:"column:is_design_approved;default:false" json:"is_design_approved"`
	IsStoreApproved    bool       `gorm:"column:is_store_approved;default:false" json:"is_store_approved"`
	TotalPrice         float64    `gorm:"column:total_price;default:0" json:"total_price"`
	VendorName         *string    `gorm:"column:vendor_name" json:"vendor_name"`
	OperatorName       *string    `gorm:"column:operator_name" json:"operator_name,omitempty"`
	OperatorReward     float64    `gorm:"column:operator_reward;default:0" json:"operator_reward"`
	SPKDate            *string    `gorm:"column:spk_date" json:"spk_date"`
	PhotoInstalled     *string    `gorm:"column:photo_installed" json:"photo_installed"`
	InstalledAt        *time.Time `gorm:"column:installed_at" json:"installed_at"`
	ValidatedAt        *time.Time `gorm:"column:validated_at" json:"validated_at"`
	EditingAt          *time.Time `gorm:"column:editing_at" json:"editing_at"`
	EditingFinishedAt  *time.Time `gorm:"column:editing_finished_at" json:"editing_finished_at"`
	OperatorReceivedAt *time.Time `gorm:"column:operator_received_at" json:"operator_received_at"`
	IsMarkedCustomer   bool       `gorm:"column:is_marked_customer;default:false" json:"is_marked_customer"`

	ValidatorID        *uuid.UUID `gorm:"type:uuid;column:validator_id" json:"validator_id,omitempty"`
	ValidatorName      *string    `gorm:"column:validator_name" json:"validator_name,omitempty"`
	ValidatorPosition  *string    `gorm:"column:validator_position" json:"validator_position,omitempty"`

	DesignValidatorID       *uuid.UUID `gorm:"type:uuid;column:design_validator_id" json:"design_validator_id,omitempty"`
	DesignValidatorName     *string    `gorm:"column:design_validator_name" json:"design_validator_name,omitempty"`
	DesignValidatorPosition *string    `gorm:"column:design_validator_position" json:"design_validator_position,omitempty"`
	DesignApprovedAt        *time.Time `gorm:"column:design_approved_at" json:"design_approved_at,omitempty"`

	StoreValidatorID       *uuid.UUID `gorm:"type:uuid;column:store_validator_id" json:"store_validator_id,omitempty"`
	StoreValidatorName     *string    `gorm:"column:store_validator_name" json:"store_validator_name,omitempty"`
	StoreValidatorPosition *string    `gorm:"column:store_validator_position" json:"store_validator_position,omitempty"`
	StoreApprovedAt        *time.Time `gorm:"column:store_approved_at" json:"store_approved_at,omitempty"`

	PurchasingID       *uuid.UUID `gorm:"type:uuid;column:purchasing_id" json:"purchasing_id,omitempty"`
	PurchasingName     *string    `gorm:"column:purchasing_name" json:"purchasing_name,omitempty"`
	PurchasingPosition *string    `gorm:"column:purchasing_position" json:"purchasing_position,omitempty"`

	IsVendorEmailSent bool       `gorm:"column:is_vendor_email_sent;default:false;index" json:"is_vendor_email_sent"`
	VendorEmailSentAt *time.Time `gorm:"column:vendor_email_sent_at" json:"vendor_email_sent_at"`

	IsPaid           bool       `gorm:"column:is_paid;default:false;index" json:"is_paid"`
	PaymentStatus    string     `gorm:"column:payment_status;type:varchar(50);default:'unpaid';index" json:"payment_status"`
	PaidAt           *time.Time `gorm:"column:paid_at" json:"paid_at"`
	PaymentReference *string    `gorm:"column:payment_reference;type:varchar(255)" json:"payment_reference"`
	PaymentNotes     *string    `gorm:"column:payment_notes;type:text" json:"payment_notes"`
	ReimbursementID  *uuid.UUID `gorm:"type:uuid;column:reimbursement_id;index" json:"reimbursement_id,omitempty"`

	Employee        *Employee       `gorm:"foreignKey:EmployeeID;constraint:OnDelete:SET NULL;" json:"employee,omitempty"`
	Designer        *Employee       `gorm:"foreignKey:DesignerID;constraint:OnDelete:SET NULL;" json:"designer,omitempty"`
	Installer       *Employee       `gorm:"foreignKey:InstallerID;constraint:OnDelete:SET NULL;" json:"installer,omitempty"`
	Editor          *Employee       `gorm:"foreignKey:EditorID;constraint:OnDelete:SET NULL;" json:"editor,omitempty"`
	Validator       *Employee       `gorm:"foreignKey:ValidatorID;constraint:OnDelete:SET NULL;" json:"validator,omitempty"`
	DesignValidator *Employee       `gorm:"foreignKey:DesignValidatorID;constraint:OnDelete:SET NULL;" json:"design_validator,omitempty"`
	StoreValidator  *Employee       `gorm:"foreignKey:StoreValidatorID;constraint:OnDelete:SET NULL;" json:"store_validator,omitempty"`
	Purchasing      *Employee       `gorm:"foreignKey:PurchasingID;constraint:OnDelete:SET NULL;" json:"purchasing,omitempty"`
	Vendor          *BrandingVendor `gorm:"foreignKey:VendorID;constraint:OnDelete:SET NULL;" json:"vendor,omitempty"`
	Operator        *Employee       `gorm:"foreignKey:OperatorID;constraint:OnDelete:SET NULL;" json:"operator,omitempty"`
}
