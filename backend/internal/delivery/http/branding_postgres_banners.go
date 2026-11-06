package http

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BannerResponse struct {
	ID                    string     `json:"id"`
	SalesID               *string    `json:"sales_id"`
	ShopID                *string    `json:"shop_id"`
	EditorID              *string    `json:"editor_id"`
	Type                  string     `json:"type"`
	BannerType            string     `json:"banner_type"`
	IsBihun               bool       `json:"is_bihun"`
	Qty                   int        `json:"qty"`
	SizeLength            float64    `json:"size_length"`
	SizeWidth             float64    `json:"size_width"`
	Size                  float64    `json:"size"`
	Notes                 *string    `json:"notes"`
	RevisionNote          *string    `json:"revision_note"`
	IsDesignApproved      bool       `json:"is_design_approved"`
	IsStoreApproved       bool       `json:"is_store_approved"`
	Status                string     `json:"status"`
	DesignLink            *string    `json:"design_link"`
	DesignURL             *string    `json:"design_url"`
	DesignFileName        *string    `json:"design_file_name"`
	TotalPrice            float64    `json:"total_price"`
	OperatorReward        float64    `json:"operator_reward"`
	VendorID              *string    `json:"vendor_id"`
	OperatorID            *string    `json:"operator_id"`
	SPKDate               *string    `json:"spk_date"`
	PhotoInstalled        *string    `json:"photo_installed"`
	InstalledAt           *time.Time `json:"installed_at"`
	CreatedAt             *time.Time `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at"`
	EditingAt             *time.Time `json:"editing_at"`
	EditingFinishedAt     *time.Time `json:"editing_finished_at"`
	ValidatedAt           *time.Time `json:"validated_at"`
	OperatorReceivedAt    *time.Time `json:"operator_received_at"`
	IsMarkedCustomer      bool       `json:"is_marked_customer"`
	DocumentationImageURL string     `json:"documentation_image_url"`
	ShopPhoto             string     `json:"shop_photo"`
	PhotoURL              string     `json:"photo_url"`
	ShopName              string     `json:"shop_name"`
	StoreName             string     `json:"store_name"`
	ShopAddress           string     `json:"shop_address"`
	Address               string     `json:"address"`
	Location              string     `json:"location"`
	Latitude              float64    `json:"latitude"`
	Longitude             float64    `json:"longitude"`
	ShopPhone             *string    `json:"shop_phone"`
	OwnerName             string     `json:"owner_name"`
	Province              *string    `json:"province"`
	Regency               *string    `json:"regency"`
	District              *string    `json:"district"`
	Village               *string    `json:"village"`
	SalesName             string     `json:"sales_name"`
	SalesPhone            *string    `json:"sales_phone"`
	SalesBranch           string     `json:"sales_branch"`
	BranchName            string     `json:"branch_name"`
	BranchID              *string    `json:"branch_id"`
	EditorName            *string    `json:"editor_name"`
	EditorUserID          *string    `json:"editor_user_id"`
	VendorName            string     `json:"vendor_name"`
	ValidatorID           *string    `json:"validator_id"`
	ValidatorName         *string    `json:"validator_name"`
	ValidatorPosition     *string    `json:"validator_position"`
	DesignValidatorID       *string    `json:"design_validator_id"`
	DesignValidatorName     *string    `json:"design_validator_name"`
	DesignValidatorPosition *string    `json:"design_validator_position"`
	DesignApprovedAt        *time.Time `json:"design_approved_at"`
	StoreValidatorID        *string    `json:"store_validator_id"`
	StoreValidatorName      *string    `json:"store_validator_name"`
	StoreValidatorPosition  *string    `json:"store_validator_position"`
	StoreApprovedAt         *time.Time `json:"store_approved_at"`
	PurchasingID          *string    `json:"purchasing_id"`
	PurchasingName        *string    `json:"purchasing_name"`
	PurchasingPosition    *string    `json:"purchasing_position"`
	IsVendorEmailSent     bool       `json:"is_vendor_email_sent"`
	VendorEmailSentAt     *time.Time `json:"vendor_email_sent_at"`
	IsPaid                bool       `json:"is_paid"`
	PaymentStatus         string     `json:"payment_status"`
	PaidAt                *time.Time `json:"paid_at"`
	PaymentReference      *string    `json:"payment_reference"`
	PaymentNotes          *string    `json:"payment_notes"`
	OperatorName          *string    `json:"operator_name"`
	OperatorPhone         *string    `json:"operator_phone"`
}

// Keep bulk banner page requests from saturating the small production database.
var bannerListSlots = make(chan struct{}, 2)

func (h *BrandingProxyHandler) GetBannersPostgres(c *gin.Context) {
	select {
	case bannerListSlots <- struct{}{}:
		defer func() { <-bannerListSlots }()
	case <-c.Request.Context().Done():
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100000 {
		limit = 20
	}
	offset := (page - 1) * limit

	statusFilter := c.Query("status")
	branchID := c.Query("branch_id")
	search := c.Query("search")
	salesName := c.Query("sales_name")
	month := c.Query("month")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	editorID := c.Query("editor_id")
	operatorID := c.Query("operator_id")
	designDate := c.Query("design_date")
	vendorID := c.Query("vendor_id")
	paymentStatus := c.Query("payment_status")
	sortBy := c.Query("sort_by")
	sortDir := strings.ToUpper(c.DefaultQuery("sort_dir", "DESC"))
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}

	onlyRevision := c.Query("only_revision")
	if onlyRevision == "" {
		onlyRevision = c.Query("has_revision")
	}

	isBihunFilter := c.Query("is_bihun")

	query := h.localDB.WithContext(c.Request.Context()).Table("banner_orders").
		Select(`
			banner_orders.id as id,
			banner_orders.employee_id::text as sales_id,
			COALESCE(banner_orders.store_id::text, stores.id::text, '') as shop_id,
			banner_orders.editor_id::text as editor_id,
			COALESCE(NULLIF(banner_orders.banner_type, ''), 'Spanduk') as type,
			COALESCE(NULLIF(banner_orders.banner_type, ''), 'Spanduk') as banner_type,
			banner_orders.is_bihun as is_bihun,
			COALESCE(banner_orders.qty, 1) as qty,
			COALESCE(banner_orders.size_length, 0) as size_length,
			COALESCE(banner_orders.size_width, 0) as size_width,
			COALESCE(NULLIF(banner_orders.size, 0), COALESCE(banner_orders.size_length, 0) * COALESCE(banner_orders.size_width, 0), 0) as size,
			banner_orders.notes as notes,
			banner_orders.revision_note as revision_note,
			COALESCE(banner_orders.is_design_approved, false) as is_design_approved,
			COALESCE(banner_orders.is_store_approved, false) as is_store_approved,
			COALESCE(banner_orders.status, 'PENDING') as status,
			COALESCE(NULLIF(banner_orders.design_link, ''), NULLIF(banner_orders.design_url, ''), '') as design_link,
			COALESCE(NULLIF(banner_orders.design_url, ''), NULLIF(banner_orders.design_link, ''), '') as design_url,
			banner_orders.design_file_name as design_file_name,
			COALESCE(banner_orders.total_price, 0) as total_price,
			COALESCE(banner_orders.operator_reward, 0) as operator_reward,
			banner_orders.vendor_id::text as vendor_id,
			COALESCE(banner_orders.vendor_name, vendors.name, '') as vendor_name,
			banner_orders.operator_id::text as operator_id,
			COALESCE(NULLIF(TRIM(CONCAT(operators.first_name, ' ', COALESCE(operators.last_name, ''))), ''), banner_orders.operator_name, '') as operator_name,
			operators.phone_number as operator_phone,
			banner_orders.spk_date as spk_date,
			banner_orders.photo_installed as photo_installed,
			banner_orders.installed_at as installed_at,
			banner_orders.created_at as created_at,
			banner_orders.updated_at as updated_at,
			banner_orders.editing_at as editing_at,
			banner_orders.editing_finished_at as editing_finished_at,
			banner_orders.validated_at as validated_at,
			banner_orders.operator_received_at as operator_received_at,
			COALESCE(banner_orders.is_marked_customer, false) as is_marked_customer,
			COALESCE(banner_orders.documentation_image_url, '') as documentation_image_url,
			COALESCE(NULLIF(stores.photo_url, ''), NULLIF(banner_orders.documentation_image_url, ''), '') as shop_photo,
			COALESCE(NULLIF(stores.photo_url, ''), NULLIF(banner_orders.documentation_image_url, ''), '') as photo_url,
			COALESCE(NULLIF(banner_orders.store_name, ''), stores.name, 'Toko (Tanpa Nama)') as shop_name,
			COALESCE(NULLIF(banner_orders.store_name, ''), stores.name, 'Toko (Tanpa Nama)') as store_name,
			COALESCE(NULLIF(NULLIF(NULLIF(stores.address, ''), 'N/A'), '-'), matched_store.address, NULLIF(NULLIF(NULLIF(banner_orders.location, ''), 'N/A'), '-'), '') as shop_address,
			COALESCE(NULLIF(NULLIF(NULLIF(stores.address, ''), 'N/A'), '-'), matched_store.address, NULLIF(NULLIF(NULLIF(banner_orders.location, ''), 'N/A'), '-'), '') as address,
			COALESCE(NULLIF(NULLIF(NULLIF(banner_orders.location, ''), 'N/A'), '-'), stores.address, matched_store.address, '') as location,
			COALESCE(stores.latitude, 0) as latitude,
			COALESCE(stores.longitude, 0) as longitude,
			COALESCE(NULLIF(stores.phone_number, ''), '') as shop_phone,
			COALESCE(stores.owner_name, '') as owner_name,
			COALESCE(stores.province, '') as province,
			COALESCE(stores.regency, '') as regency,
			COALESCE(stores.district, '') as district,
			COALESCE(stores.village, '') as village,
			COALESCE(banner_orders.mysql_sales_name, NULLIF(TRIM(CONCAT(sales.first_name, ' ', COALESCE(sales.last_name, ''))), ''), 'Unknown Sales') as sales_name,
			sales.phone_number as sales_phone,
			COALESCE(branches.name, op_branches.name, store_branches.name, store_op_branches.name, '') as sales_branch,
			COALESCE(branches.name, op_branches.name, store_branches.name, store_op_branches.name, '') as branch_name,
			COALESCE(sales.branch_id::text, sales.operating_branch_id::text, store_sales.branch_id::text, store_sales.operating_branch_id::text, '') as branch_id,
			COALESCE(NULLIF(TRIM(CONCAT(editors.first_name, ' ', COALESCE(editors.last_name, ''))), ''), banner_orders.mysql_editor_name, '') as editor_name,
			editors.user_id::text as editor_user_id,
			banner_orders.validator_id::text as validator_id,
			COALESCE(banner_orders.validator_name, NULLIF(TRIM(CONCAT(validators.first_name, ' ', COALESCE(validators.last_name, ''))), ''), '') as validator_name,
			COALESCE(banner_orders.validator_position, validator_positions.title, '') as validator_position,
			banner_orders.design_validator_id::text as design_validator_id,
			COALESCE(banner_orders.design_validator_name, NULLIF(TRIM(CONCAT(design_validators.first_name, ' ', COALESCE(design_validators.last_name, ''))), ''), banner_orders.validator_name, '') as design_validator_name,
			COALESCE(banner_orders.design_validator_position, design_validator_positions.title, banner_orders.validator_position, '') as design_validator_position,
			banner_orders.design_approved_at as design_approved_at,
			banner_orders.store_validator_id::text as store_validator_id,
			COALESCE(banner_orders.store_validator_name, NULLIF(TRIM(CONCAT(store_validators.first_name, ' ', COALESCE(store_validators.last_name, ''))), ''), banner_orders.validator_name, '') as store_validator_name,
			COALESCE(banner_orders.store_validator_position, store_validator_positions.title, banner_orders.validator_position, '') as store_validator_position,
			banner_orders.store_approved_at as store_approved_at,
			banner_orders.purchasing_id::text as purchasing_id,
			COALESCE(banner_orders.purchasing_name, NULLIF(TRIM(CONCAT(purchasings.first_name, ' ', COALESCE(purchasings.last_name, ''))), ''), '') as purchasing_name,
			COALESCE(banner_orders.purchasing_position, purchasing_positions.title, '') as purchasing_position,
			COALESCE(banner_orders.is_vendor_email_sent, false) as is_vendor_email_sent,
			banner_orders.vendor_email_sent_at as vendor_email_sent_at,
			COALESCE(banner_orders.is_paid, false) as is_paid,
			COALESCE(banner_orders.payment_status, 'unpaid') as payment_status,
			banner_orders.paid_at as paid_at,
			banner_orders.payment_reference as payment_reference,
			banner_orders.payment_notes as payment_notes
		`).
		Joins("LEFT JOIN stores ON stores.id = banner_orders.store_id").
		Joins(`LEFT JOIN LATERAL (
			SELECT MAX(NULLIF(TRIM(s.address), '')) AS address
			FROM stores s
			WHERE banner_orders.store_id IS NULL
			  AND s.company_id = banner_orders.company_id
			  AND LOWER(TRIM(s.name)) = LOWER(TRIM(banner_orders.store_name))
			HAVING COUNT(*) = 1
		) matched_store ON TRUE`).
		Joins("LEFT JOIN employees sales ON sales.id = banner_orders.employee_id").
		Joins("LEFT JOIN branches ON branches.id = sales.branch_id").
		Joins("LEFT JOIN branches op_branches ON op_branches.id = sales.operating_branch_id").
		Joins("LEFT JOIN employees store_sales ON store_sales.id = stores.assigned_employee_id").
		Joins("LEFT JOIN branches store_branches ON store_branches.id = store_sales.branch_id").
		Joins("LEFT JOIN branches store_op_branches ON store_op_branches.id = store_sales.operating_branch_id").
		Joins("LEFT JOIN branding_vendors vendors ON vendors.id = banner_orders.vendor_id").
		Joins("LEFT JOIN employees editors ON editors.id = banner_orders.editor_id").
		Joins("LEFT JOIN employees operators ON operators.id = banner_orders.operator_id").
		Joins("LEFT JOIN employees validators ON validators.id = banner_orders.validator_id").
		Joins("LEFT JOIN job_positions validator_positions ON validator_positions.id = validators.job_position_id").
		Joins("LEFT JOIN employees design_validators ON design_validators.id = banner_orders.design_validator_id").
		Joins("LEFT JOIN job_positions design_validator_positions ON design_validator_positions.id = design_validators.job_position_id").
		Joins("LEFT JOIN employees store_validators ON store_validators.id = banner_orders.store_validator_id").
		Joins("LEFT JOIN job_positions store_validator_positions ON store_validator_positions.id = store_validators.job_position_id").
		Joins("LEFT JOIN employees purchasings ON purchasings.id = banner_orders.purchasing_id").
		Joins("LEFT JOIN job_positions purchasing_positions ON purchasing_positions.id = purchasings.job_position_id")

	applyStatusFilter := func(q *gorm.DB) *gorm.DB {
		if statusFilter == "" || statusFilter == "all" {
			return q
		}
		switch statusFilter {
		case "request":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{"request", "pending"})
		case "editing":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{"editing", "in_design"})
		case "validation":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{"validation", "val_design", "val_store", "admin_review"})
		case "purchasing_queue":
			return q.Where("LOWER(banner_orders.status) = 'purchasing_queue'")
		case "sent_to_vendor", "vendor", "percetakan":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{"sent_to_vendor", "percetakan", "printed", "in_production"})
		case "ready_for_operator", "siap_pasang", "siap pasang":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{
				"ready_for_operator", "diterima_operator", "siap_pasang", "siap pasang", "siap-pasang", "sent_to_vendor", "percetakan", "printed", "in_production",
			})
		case "installation":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{
				"ready_for_operator", "diterima_operator", "siap_pasang", "siap pasang", "siap-pasang", "sent_to_vendor", "percetakan", "printed", "in_production",
			})
		case "installed":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{"installed", "terpasang", "selesai", "selesai_dipasang"})
		case "done":
			return q.Where("LOWER(banner_orders.status) IN ?", []string{
				"sent_to_vendor", "purchasing_queue", "percetakan", "printed", "in_production", "ready_for_operator", "diterima_operator", "installed", "terpasang", "selesai", "selesai_dipasang", "siap_pasang", "siap pasang",
			})
		default:
			if strings.Contains(statusFilter, ",") {
				statuses := strings.Split(statusFilter, ",")
				var lowerStatuses []string
				for _, s := range statuses {
					lowerStatuses = append(lowerStatuses, strings.ToLower(strings.TrimSpace(s)))
				}
				return q.Where("LOWER(banner_orders.status) IN ?", lowerStatuses)
			}
			return q.Where("LOWER(banner_orders.status) = ?", strings.ToLower(statusFilter))
		}
	}

	query = applyStatusFilter(query)

	if onlyRevision == "true" || onlyRevision == "1" {
		query = query.Where("banner_orders.revision_note IS NOT NULL AND banner_orders.revision_note != ''")
	}

	if paymentStatus == "unpaid" {
		query = query.Where("(banner_orders.is_paid = false OR banner_orders.payment_status = 'unpaid' OR banner_orders.payment_status IS NULL)")
	} else if paymentStatus == "paid" {
		query = query.Where("(banner_orders.is_paid = true OR banner_orders.payment_status = 'paid')")
	}

	if vendorID != "" && vendorID != "all" {
		query = query.Where("banner_orders.vendor_id::text = ?", vendorID)
	}

	applySearchFilter := func(db *gorm.DB) *gorm.DB {
		if search == "" {
			return db
		}
		trimmed := strings.TrimSpace(search)
		searchQuery := "%" + strings.ToLower(trimmed) + "%"

		return db.Where(`(
			LOWER(COALESCE(banner_orders.store_name, '')) LIKE ? OR 
			LOWER(COALESCE(stores.name, '')) LIKE ? OR
			LOWER(COALESCE(stores.owner_name, '')) LIKE ? OR
			LOWER(COALESCE(stores.address, '')) LIKE ? OR
			LOWER(COALESCE(banner_orders.location, '')) LIKE ? OR
			LOWER(COALESCE(stores.district, '')) LIKE ? OR
			LOWER(COALESCE(stores.regency, '')) LIKE ? OR
			LOWER(COALESCE(stores.village, '')) LIKE ? OR
			LOWER(COALESCE(branches.name, '')) LIKE ? OR
			LOWER(COALESCE(op_branches.name, '')) LIKE ? OR
			LOWER(COALESCE(banner_orders.notes, '')) LIKE ? OR
			LOWER(COALESCE(banner_orders.banner_type, '')) LIKE ? OR
			LOWER(COALESCE(banner_orders.mysql_editor_name, '')) LIKE ? OR
			LOWER(COALESCE(editors.first_name, '')) LIKE ? OR
			LOWER(COALESCE(editors.last_name, '')) LIKE ? OR
			LOWER(TRIM(CONCAT(COALESCE(editors.first_name, ''), ' ', COALESCE(editors.last_name, '')))) LIKE ? OR
			LOWER(COALESCE(banner_orders.mysql_sales_name, '')) LIKE ? OR
			LOWER(COALESCE(sales.first_name, '')) LIKE ? OR
			LOWER(COALESCE(sales.last_name, '')) LIKE ? OR
			LOWER(TRIM(CONCAT(COALESCE(sales.first_name, ''), ' ', COALESCE(sales.last_name, '')))) LIKE ? OR
			LOWER(COALESCE(banner_orders.operator_name, '')) LIKE ? OR
			LOWER(COALESCE(operators.first_name, '')) LIKE ? OR
			LOWER(COALESCE(operators.last_name, '')) LIKE ? OR
			LOWER(TRIM(CONCAT(COALESCE(operators.first_name, ''), ' ', COALESCE(operators.last_name, '')))) LIKE ?
		)`, 
			searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery,
			searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery, searchQuery,
			searchQuery, searchQuery, searchQuery, searchQuery,
		)
	}

	query = applySearchFilter(query)

	if branchID != "" && branchID != "ALL_BRANCHES" {
		query = query.Where("(sales.branch_id::text = ? OR sales.operating_branch_id::text = ? OR store_sales.branch_id::text = ? OR store_sales.operating_branch_id::text = ?)", branchID, branchID, branchID, branchID)
	}

	if salesName != "" && salesName != "all" {
		sName := "%" + strings.ToLower(salesName) + "%"
		query = query.Where("(LOWER(COALESCE(banner_orders.mysql_sales_name, '')) LIKE ? OR LOWER(TRIM(CONCAT(sales.first_name, ' ', COALESCE(sales.last_name, '')))) LIKE ?)", sName, sName)
	}

	if editorID == "unassigned" || editorID == "none" {
		query = query.Where("banner_orders.editor_id IS NULL AND (banner_orders.mysql_editor_name IS NULL OR banner_orders.mysql_editor_name = '')")
	} else if editorID != "" && editorID != "all" {
		eFilter := "%" + strings.ToLower(editorID) + "%"
		query = query.Where("(banner_orders.editor_id::text = ? OR LOWER(COALESCE(banner_orders.mysql_editor_name, '')) LIKE ? OR LOWER(COALESCE(editors.first_name, '')) LIKE ?)", editorID, eFilter, eFilter)
	}

	if operatorID == "unassigned" || operatorID == "none" {
		query = query.Where("banner_orders.operator_id IS NULL AND (banner_orders.operator_name IS NULL OR banner_orders.operator_name = '')")
	} else if operatorID != "" && operatorID != "all" {
		opFilter := "%" + strings.ToLower(operatorID) + "%"
		query = query.Where("(banner_orders.operator_id::text = ? OR LOWER(COALESCE(banner_orders.operator_name, '')) LIKE ? OR LOWER(COALESCE(operators.first_name, '')) LIKE ? OR LOWER(COALESCE(operators.last_name, '')) LIKE ? OR LOWER(TRIM(CONCAT(COALESCE(operators.first_name, ''), ' ', COALESCE(operators.last_name, '')))) LIKE ?)", operatorID, opFilter, opFilter, opFilter, opFilter)
	}

	if designDate != "" && designDate != "all" {
		query = query.Where("TO_CHAR(COALESCE(banner_orders.editing_finished_at, banner_orders.editing_at), 'YYYY-MM-DD') = ?", designDate)
	}

	if month != "" && month != "all" {
		query = query.Where("TO_CHAR(COALESCE(banner_orders.installed_at, banner_orders.created_at), 'YYYY-MM') = ?", month)
	}

	if startDate != "" {
		query = query.Where("DATE(COALESCE(banner_orders.installed_at, banner_orders.created_at)) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(COALESCE(banner_orders.installed_at, banner_orders.created_at)) <= ?", endDate)
	}

	if isBihunFilter == "true" || isBihunFilter == "1" {
		query = query.Where("banner_orders.is_bihun = true")
	} else if isBihunFilter == "false" || isBihunFilter == "0" {
		query = query.Where("(banner_orders.is_bihun = false OR banner_orders.is_bihun IS NULL)")
	}

	var totalCount int64
	// Fast count query: only join strictly required tables
	countQuery := h.localDB.WithContext(c.Request.Context()).Table("banner_orders")
	needsStoreJoin := search != "" || (branchID != "" && branchID != "ALL_BRANCHES")
	needsSalesJoin := search != "" || (salesName != "" && salesName != "all") || (branchID != "" && branchID != "ALL_BRANCHES")
	needsEditorJoin := search != "" || (editorID != "" && editorID != "all")
	needsOperatorJoin := search != "" || (operatorID != "" && operatorID != "all")

	if needsStoreJoin {
		countQuery = countQuery.Joins("LEFT JOIN stores ON stores.id = banner_orders.store_id")
	}
	if needsSalesJoin {
		countQuery = countQuery.Joins("LEFT JOIN employees sales ON sales.id = banner_orders.employee_id")
		countQuery = countQuery.Joins("LEFT JOIN branches branches ON branches.id = sales.branch_id")
		countQuery = countQuery.Joins("LEFT JOIN branches op_branches ON op_branches.id = sales.operating_branch_id")
		if branchID != "" && branchID != "ALL_BRANCHES" {
			countQuery = countQuery.Joins("LEFT JOIN employees store_sales ON store_sales.id = stores.assigned_employee_id")
		}
	}
	if needsEditorJoin {
		countQuery = countQuery.Joins("LEFT JOIN employees editors ON editors.id = banner_orders.editor_id")
	}
	if needsOperatorJoin {
		countQuery = countQuery.Joins("LEFT JOIN employees operators ON operators.id = banner_orders.operator_id")
	}

	countQuery = applySearchFilter(countQuery)

	if branchID != "" && branchID != "ALL_BRANCHES" {
		countQuery = countQuery.Where("(sales.branch_id::text = ? OR sales.operating_branch_id::text = ? OR store_sales.branch_id::text = ? OR store_sales.operating_branch_id::text = ?)", branchID, branchID, branchID, branchID)
	}

	if salesName != "" && salesName != "all" {
		sName := "%" + strings.ToLower(salesName) + "%"
		countQuery = countQuery.Where("(LOWER(COALESCE(banner_orders.mysql_sales_name, '')) LIKE ? OR LOWER(TRIM(CONCAT(sales.first_name, ' ', COALESCE(sales.last_name, '')))) LIKE ?)", sName, sName)
	}

	if editorID == "unassigned" || editorID == "none" {
		countQuery = countQuery.Where("banner_orders.editor_id IS NULL AND (banner_orders.mysql_editor_name IS NULL OR banner_orders.mysql_editor_name = '')")
	} else if editorID != "" && editorID != "all" {
		eFilter := "%" + strings.ToLower(editorID) + "%"
		countQuery = countQuery.Where("(banner_orders.editor_id::text = ? OR LOWER(COALESCE(banner_orders.mysql_editor_name, '')) LIKE ? OR LOWER(COALESCE(editors.first_name, '')) LIKE ?)", editorID, eFilter, eFilter)
	}

	if operatorID == "unassigned" || operatorID == "none" {
		countQuery = countQuery.Where("banner_orders.operator_id IS NULL AND (banner_orders.operator_name IS NULL OR banner_orders.operator_name = '')")
	} else if operatorID != "" && operatorID != "all" {
		opFilter := "%" + strings.ToLower(operatorID) + "%"
		countQuery = countQuery.Where("(banner_orders.operator_id::text = ? OR LOWER(COALESCE(banner_orders.operator_name, '')) LIKE ? OR LOWER(COALESCE(operators.first_name, '')) LIKE ? OR LOWER(COALESCE(operators.last_name, '')) LIKE ? OR LOWER(TRIM(CONCAT(COALESCE(operators.first_name, ''), ' ', COALESCE(operators.last_name, '')))) LIKE ?)", operatorID, opFilter, opFilter, opFilter, opFilter)
	}

	if onlyRevision == "true" || onlyRevision == "1" {
		countQuery = countQuery.Where("banner_orders.revision_note IS NOT NULL AND banner_orders.revision_note != ''")
	}

	countQuery = applyStatusFilter(countQuery)

	if paymentStatus == "unpaid" {
		countQuery = countQuery.Where("(banner_orders.is_paid = false OR banner_orders.payment_status = 'unpaid' OR banner_orders.payment_status IS NULL)")
	} else if paymentStatus == "paid" {
		countQuery = countQuery.Where("(banner_orders.is_paid = true OR banner_orders.payment_status = 'paid')")
	}

	if vendorID != "" && vendorID != "all" {
		countQuery = countQuery.Where("banner_orders.vendor_id::text = ?", vendorID)
	}

	if designDate != "" && designDate != "all" {
		countQuery = countQuery.Where("TO_CHAR(COALESCE(banner_orders.editing_finished_at, banner_orders.editing_at), 'YYYY-MM-DD') = ?", designDate)
	}

	if month != "" && month != "all" {
		countQuery = countQuery.Where("TO_CHAR(COALESCE(banner_orders.installed_at, banner_orders.created_at), 'YYYY-MM') = ?", month)
	}

	if startDate != "" {
		countQuery = countQuery.Where("DATE(COALESCE(banner_orders.installed_at, banner_orders.created_at)) >= ?", startDate)
	}
	if endDate != "" {
		countQuery = countQuery.Where("DATE(COALESCE(banner_orders.installed_at, banner_orders.created_at)) <= ?", endDate)
	}

	if isBihunFilter == "true" || isBihunFilter == "1" {
		countQuery = countQuery.Where("banner_orders.is_bihun = true")
	} else if isBihunFilter == "false" || isBihunFilter == "0" {
		countQuery = countQuery.Where("(banner_orders.is_bihun = false OR banner_orders.is_bihun IS NULL)")
	}

	if err := countQuery.Count(&totalCount).Error; err != nil {
		log.Printf("[BrandingDB] PostgreSQL Count error: %v", err)
	}

	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset = (page - 1) * limit
	isUnlimited := limitStr == "99999" || limitStr == "-1" || limitStr == "all"

	switch sortBy {
	case "editor_name", "editor":
		query = query.Order("COALESCE(NULLIF(TRIM(CONCAT(editors.first_name, ' ', COALESCE(editors.last_name, ''))), ''), banner_orders.mysql_editor_name, '') " + sortDir + " NULLS LAST, banner_orders.created_at DESC")
	case "shop_name", "store_name", "store", "name":
		query = query.Order("banner_orders.store_name " + sortDir + ", banner_orders.created_at DESC")
	case "revision", "revision_note":
		query = query.Order("CASE WHEN banner_orders.revision_note IS NOT NULL AND banner_orders.revision_note != '' THEN 1 ELSE 0 END DESC, banner_orders.updated_at " + sortDir + ", banner_orders.created_at DESC")
	case "editing_finished_at":
		query = query.Order("banner_orders.editing_finished_at " + sortDir + " NULLS LAST, banner_orders.created_at DESC")
	case "installed_at":
		query = query.Order("banner_orders.installed_at " + sortDir + " NULLS LAST, banner_orders.created_at DESC")
	case "updated_at", "latest_processed", "sent_to_vendor", "process_date":
		query = query.Order("banner_orders.updated_at " + sortDir + ", banner_orders.created_at DESC")
	case "paid_at":
		query = query.Order("banner_orders.paid_at " + sortDir + " NULLS LAST, banner_orders.updated_at DESC")
	case "created_at", "date":
		query = query.Order("banner_orders.created_at " + sortDir)
	default:
		if sortBy == "" && (statusFilter == "sent_to_vendor" || statusFilter == "vendor" || statusFilter == "percetakan") {
			query = query.Order("banner_orders.updated_at DESC, banner_orders.created_at DESC")
		} else {
			query = query.Order("banner_orders.created_at " + sortDir)
		}
	}

	if !isUnlimited {
		query = query.Offset(offset).Limit(limit)
	}

	var banners []BannerResponse
	if err := query.Find(&banners).Error; err != nil {
		log.Printf("[BrandingDB] PostgreSQL GetBanners error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil data dari PostgreSQL",
			"details": err.Error(),
		})
		return
	}

	total := totalCount
	if total == 0 && len(banners) > 0 {
		total = int64(len(banners))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"source":  "postgres",
		"data":    banners,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
