package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleSuperAdmin       Role = "superadmin"
	RoleHRAdmin          Role = "hr_admin"
	RoleEmployee         Role = "employee"
	RoleManager          Role = "manager"
	RoleAdminNota        Role = "admin_nota"
	RoleSupervisorSales    Role = "supervisor_sales"
	RoleSupervisorSalesMT  Role = "supervisor_sales_mt"
	RoleCashier            Role = "cashier"
	RoleCashierPusat     Role = "cashier_pusat"
	RoleAdminKasir       Role = "admin_kasir"
	RoleAdminGudang      Role = "admin_gudang"
	RoleAuditor          Role = "auditor"
	RoleAdminDesigner    Role = "admin_designer"
	RoleValidatorDesain  Role = "validator_desain"
	RoleValidatorToko    Role = "validator_toko"
	RoleAdminPurchasing  Role = "admin_purchasing"
	RoleOperator         Role = "operator"
	RoleCustomerSupport  Role = "customer_support"
	RoleAdminMarketplace Role = "admin_marketplace"
	RoleAdminOperasional Role = "admin_operasional"
	RoleAdminPabrik      Role = "admin_pabrik"
	RoleSupervisorGudang Role = "supervisor_gudang"
	RoleDistributorAdmin Role = "distributor_admin"
	RoleDistributorSales Role = "distributor_sales"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string     `gorm:"type:varchar(255);unique;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Role         Role       `gorm:"type:varchar(50);default:'employee'" json:"role"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"default:now()" json:"updated_at"`
	CompanyID    *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	AccessRoleID *uuid.UUID `gorm:"type:uuid" json:"access_role_id"`
	CanAbsenAnywhere bool `gorm:"default:false" json:"can_absen_anywhere"`
	MobileSessionID  *string    `gorm:"type:varchar(64)" json:"mobile_session_id,omitempty"`
	DistributorStoreID *uuid.UUID `gorm:"type:uuid" json:"distributor_store_id,omitempty"`

	Company          *Company    `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	AccessRole       *AccessRole `gorm:"foreignKey:AccessRoleID" json:"access_role,omitempty"`
	Employee         *Employee   `gorm:"foreignKey:UserID" json:"employee,omitempty"`
	DistributorStore *Store      `gorm:"foreignKey:DistributorStoreID" json:"distributor_store,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
