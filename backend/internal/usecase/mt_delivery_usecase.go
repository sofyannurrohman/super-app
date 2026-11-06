package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type MTDeliveryUsecase interface {
	// Factory/Pusat: membuat dan mengelola DO
	CreateDO(req CreateMTDeliveryRequest, companyID uuid.UUID, createdBy string) (*domain.MTDeliveryOrder, error)
	UpdateDO(id uuid.UUID, req CreateMTDeliveryRequest) (*domain.MTDeliveryOrder, error)
	AssignToBranch(id uuid.UUID, branchID uuid.UUID) error
	CancelDO(id uuid.UUID, reason string) error
	DeleteDO(id uuid.UUID) error

	// Factory: List semua DO perusahaan
	ListByCompany(companyID uuid.UUID) ([]domain.MTDeliveryOrder, error)
	GetByID(id uuid.UUID) (*domain.MTDeliveryOrder, error)

	// Branch/Cabang: approve, siapkan, muat, kirim
	ListByBranch(branchID uuid.UUID) ([]domain.MTDeliveryOrder, error)
	ApproveDO(id uuid.UUID, approvedBy string) error
	RejectDO(id uuid.UUID, reason string) error
	StartLoading(id uuid.UUID, preparedBy string, items []PreparedItemRequest) error
	ShipDO(id uuid.UUID, vehicleID, driverID *uuid.UUID, notes string) error
	MarkDelivered(id uuid.UUID) error
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type MTDeliveryItemRequest struct {
	ProductID    uuid.UUID `json:"product_id" binding:"required"`
	RequestedQty int       `json:"requested_qty" binding:"required,min=1"`
	Unit         string    `json:"unit"`
	BatchNo      string    `json:"batch_no"`
	Notes        string    `json:"notes"`
	// Harga satuan untuk SO line-item (opsional, default 0)
	UnitPrice float64 `json:"unit_price"`
}

type CreateMTDeliveryRequest struct {
	KeyAccountID    uuid.UUID               `json:"key_account_id" binding:"required"`
	PONumber        string                  `json:"po_number"`
	SourceType      string                  `json:"source_type"` // FACTORY, BRANCH, MIXED
	SourceFactoryID *uuid.UUID              `json:"source_factory_id"`
	SourceBranchID  *uuid.UUID              `json:"source_branch_id"`
	DCDestination   string                  `json:"dc_destination"`
	ScheduledDate   *time.Time              `json:"scheduled_date,omitempty"`
	Notes           string                  `json:"notes"`
	Items           []MTDeliveryItemRequest `json:"items"`

	// ── Auto-create Sales Order ───────────────────────────────────────────
	// SalesmanEmployeeID: ID karyawan Sales MT yang bertanggung jawab atas DO ini.
	// Jika diisi, sistem akan otomatis membuat SalesOrder atas nama salesman ini.
	SalesmanEmployeeID *uuid.UUID `json:"salesman_employee_id"`
	// SalesmanBranchID: cabang tempat salesman bertugas (wajib jika SalesmanEmployeeID diisi)
	SalesmanBranchID *uuid.UUID `json:"salesman_branch_id"`
	// MTStoreID: ID toko/outlet MT sebagai tujuan SO (wajib jika SalesmanEmployeeID diisi)
	MTStoreID *uuid.UUID `json:"mt_store_id"`
}

type PreparedItemRequest struct {
	ProductID   uuid.UUID `json:"product_id"`
	PreparedQty int       `json:"prepared_qty"`
	ShippedQty  int       `json:"shipped_qty"`
	BatchNo     string    `json:"batch_no"`
	Notes       string    `json:"notes"`
}

// ─── Implementation ───────────────────────────────────────────────────────────

type mtDeliveryUsecase struct {
	repo      repository.MTDeliveryRepository
	db        *gorm.DB
	fcmClient FcmUsecase
}

func NewMTDeliveryUsecase(repo repository.MTDeliveryRepository, db *gorm.DB, fcmClient FcmUsecase) MTDeliveryUsecase {
	return &mtDeliveryUsecase{repo: repo, db: db, fcmClient: fcmClient}
}

// generateSONumber menghasilkan nomor SO unik untuk MT: SOMT-YYYYMMDD-NNNN
func (u *mtDeliveryUsecase) generateSONumber() string {
	now := time.Now()
	prefix := fmt.Sprintf("SOMT-%s-", now.Format("20060102"))
	var count int64
	u.db.Model(&domain.SalesOrder{}).
		Where("so_number LIKE ?", prefix+"%").
		Count(&count)
	return fmt.Sprintf("%s%04d", prefix, count+1)
}

func (u *mtDeliveryUsecase) CreateDO(req CreateMTDeliveryRequest, companyID uuid.UUID, createdBy string) (*domain.MTDeliveryOrder, error) {
	doNumber := u.repo.GenerateDONumber(companyID)
	sourceType := domain.MTDeliverySourceType(req.SourceType)
	if sourceType == "" {
		sourceType = domain.MTDeliverySourceBranch
	}

	items := make([]domain.MTDeliveryOrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		unit := it.Unit
		if unit == "" {
			unit = "pcs"
		}
		batchNo := it.BatchNo
		if batchNo == "" {
			batchNo = "DEFAULT"
		}
		items = append(items, domain.MTDeliveryOrderItem{
			ProductID:    it.ProductID,
			RequestedQty: it.RequestedQty,
			Unit:         unit,
			BatchNo:      batchNo,
			Notes:        it.Notes,
		})
	}

	order := &domain.MTDeliveryOrder{
		CompanyID:       companyID,
		DONumber:        doNumber,
		KeyAccountID:    req.KeyAccountID,
		PONumber:        req.PONumber,
		SourceType:      sourceType,
		SourceFactoryID: req.SourceFactoryID,
		SourceBranchID:  req.SourceBranchID,
		Status:          domain.MTDeliveryDraft,
		DCDestination:   req.DCDestination,
		ScheduledDate:   req.ScheduledDate,
		Notes:           req.Notes,
		CreatedBy:       createdBy,
		Items:           items,
	}

	// ── Auto-create Sales Order jika SalesmanEmployeeID disediakan ────────
	if req.SalesmanEmployeeID != nil {
		if req.SalesmanBranchID == nil {
			return nil, errors.New("salesman_branch_id wajib diisi jika salesman_employee_id diisi")
		}
		if req.MTStoreID == nil {
			return nil, errors.New("mt_store_id wajib diisi jika salesman_employee_id diisi")
		}

		so, err := u.createLinkedSalesOrder(req, companyID, createdBy)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat Sales Order terkait: %w", err)
		}
		order.SalesOrderID = &so.ID
	}

	if err := u.repo.Create(order); err != nil {
		return nil, err
	}
	return order, nil
}

// createLinkedSalesOrder membuat SalesOrder baru yang terhubung ke MTDeliveryOrder.
// SO ini berperan sebagai "pesanan" internal dari Sales MT ke pabrik,
// dan akan melewati alur approval Admin Nota → pengiriman → faktur → kasir.
func (u *mtDeliveryUsecase) createLinkedSalesOrder(req CreateMTDeliveryRequest, companyID uuid.UUID, createdBy string) (*domain.SalesOrder, error) {
	soNumber := u.generateSONumber()

	// Hitung total amount dari items
	var totalAmount float64
	soItems := make([]domain.SalesOrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		unit := it.Unit
		if unit == "" {
			unit = "PCS"
		}
		lineTotal := it.UnitPrice * float64(it.RequestedQty)
		totalAmount += lineTotal
		soItems = append(soItems, domain.SalesOrderItem{
			ProductID:       it.ProductID,
			Quantity:        it.RequestedQty,
			OrderedQuantity: it.RequestedQty,
			Unit:            unit,
			PiecesPerUnit:   1,
			Price:           it.UnitPrice,
			Subtotal:        lineTotal,
		})
	}

	notes := req.Notes
	if notes == "" {
		notes = fmt.Sprintf("Auto-generated dari MT DO (PO: %s, DC: %s)", req.PONumber, req.DCDestination)
	}

	so := &domain.SalesOrder{
		CompanyID:     companyID,
		BranchID:      *req.SalesmanBranchID,
		EmployeeID:    *req.SalesmanEmployeeID,
		StoreID:       *req.MTStoreID,
		StoreCategory: "MT",
		SONumber:      soNumber,
		Status:        domain.SOStatusDraft,
		TotalAmount:   totalAmount,
		Notes:         notes,
		OrderDate:     time.Now(),
		Items:         soItems,
		// MT flag: faktur dan omzet SO ini dikelola oleh pusat/pabrik
		IsMT:             true,
		FactoryCompanyID: &companyID,
	}

	if err := u.db.Create(so).Error; err != nil {
		return nil, err
	}
	return so, nil
}

func (u *mtDeliveryUsecase) UpdateDO(id uuid.UUID, req CreateMTDeliveryRequest) (*domain.MTDeliveryOrder, error) {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.MTDeliveryDraft && order.Status != domain.MTDeliveryAssigned {
		return nil, errors.New("hanya DO dengan status DRAFT atau ASSIGNED yang bisa diedit")
	}

	items := make([]domain.MTDeliveryOrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		unit := it.Unit
		if unit == "" {
			unit = "pcs"
		}
		batchNo := it.BatchNo
		if batchNo == "" {
			batchNo = "DEFAULT"
		}
		items = append(items, domain.MTDeliveryOrderItem{
			MTDeliveryOrderID: id,
			ProductID:         it.ProductID,
			RequestedQty:      it.RequestedQty,
			Unit:              unit,
			BatchNo:           batchNo,
			Notes:             it.Notes,
		})
	}

	order.KeyAccountID = req.KeyAccountID
	order.PONumber = req.PONumber
	if req.SourceType != "" {
		order.SourceType = domain.MTDeliverySourceType(req.SourceType)
	}
	order.SourceFactoryID = req.SourceFactoryID
	order.SourceBranchID = req.SourceBranchID
	order.DCDestination = req.DCDestination
	order.ScheduledDate = req.ScheduledDate
	order.Notes = req.Notes
	order.Items = items

	if err := u.repo.Update(order); err != nil {
		return nil, err
	}
	return order, nil
}

func (u *mtDeliveryUsecase) AssignToBranch(id uuid.UUID, branchID uuid.UUID) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != domain.MTDeliveryDraft {
		return errors.New("hanya DO berstatus DRAFT yang bisa di-assign ke cabang")
	}
	order.SourceBranchID = &branchID
	order.Status = domain.MTDeliveryAssigned
	order.Items = nil // Jangan timpa items saat update status
	return u.repo.Update(order)
}

func (u *mtDeliveryUsecase) CancelDO(id uuid.UUID, reason string) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status == domain.MTDeliveryShipped || order.Status == domain.MTDeliveryDelivered {
		return errors.New("DO yang sudah dikirim tidak bisa dibatalkan")
	}
	order.Status = domain.MTDeliveryCancelled
	order.RejectionReason = reason
	order.Items = nil
	return u.repo.Update(order)
}

func (u *mtDeliveryUsecase) DeleteDO(id uuid.UUID) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != domain.MTDeliveryDraft && order.Status != domain.MTDeliveryCancelled {
		return errors.New("hanya DO berstatus DRAFT atau CANCELLED yang bisa dihapus")
	}
	return u.repo.Delete(id)
}

func (u *mtDeliveryUsecase) resolveBranchNames(orders []domain.MTDeliveryOrder) []domain.MTDeliveryOrder {
	var branchIDs []string
	for _, order := range orders {
		if order.DCDestination != "" {
			if _, err := uuid.Parse(order.DCDestination); err == nil {
				branchIDs = append(branchIDs, order.DCDestination)
			}
		}
	}
	if len(branchIDs) > 0 {
		var branches []domain.Branch
		if err := u.db.Where("id IN ?", branchIDs).Find(&branches).Error; err == nil {
			branchMap := make(map[string]string)
			for _, b := range branches {
				branchMap[b.ID.String()] = b.Name
			}
			for i, order := range orders {
				if name, ok := branchMap[order.DCDestination]; ok {
					orders[i].DCDestination = name
				}
			}
		}
	}
	return orders
}

func (u *mtDeliveryUsecase) ListByCompany(companyID uuid.UUID) ([]domain.MTDeliveryOrder, error) {
	orders, err := u.repo.ListByCompany(companyID)
	if err != nil {
		return nil, err
	}
	return u.resolveBranchNames(orders), nil
}

func (u *mtDeliveryUsecase) GetByID(id uuid.UUID) (*domain.MTDeliveryOrder, error) {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order != nil && order.DCDestination != "" {
		if _, err := uuid.Parse(order.DCDestination); err == nil {
			var branch domain.Branch
			if err := u.db.First(&branch, "id = ?", order.DCDestination).Error; err == nil {
				order.DCDestination = branch.Name
			}
		}
	}
	return order, nil
}

func (u *mtDeliveryUsecase) ListByBranch(branchID uuid.UUID) ([]domain.MTDeliveryOrder, error) {
	orders, err := u.repo.ListByBranch(branchID)
	if err != nil {
		return nil, err
	}
	return u.resolveBranchNames(orders), nil
}

func (u *mtDeliveryUsecase) ApproveDO(id uuid.UUID, approvedBy string) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != domain.MTDeliveryAssigned {
		return errors.New("hanya DO berstatus ASSIGNED yang bisa di-approve")
	}
	order.Status = domain.MTDeliveryApproved
	order.ApprovedBy = approvedBy
	order.Items = nil
	return u.repo.Update(order)
}

func (u *mtDeliveryUsecase) RejectDO(id uuid.UUID, reason string) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != domain.MTDeliveryAssigned {
		return errors.New("hanya DO berstatus ASSIGNED yang bisa ditolak")
	}
	order.Status = domain.MTDeliveryRejected
	order.RejectionReason = reason
	order.Items = nil
	return u.repo.Update(order)
}

func (u *mtDeliveryUsecase) StartLoading(id uuid.UUID, preparedBy string, items []PreparedItemRequest) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != domain.MTDeliveryApproved {
		return errors.New("hanya DO berstatus APPROVED yang bisa dimulai pemuatannya")
	}

	// Update prepared/shipped qty per item
	for i := range order.Items {
		for _, prep := range items {
			if order.Items[i].ProductID == prep.ProductID {
				order.Items[i].PreparedQty = prep.PreparedQty
				order.Items[i].ShippedQty = prep.ShippedQty
				if prep.BatchNo != "" {
					order.Items[i].BatchNo = prep.BatchNo
				}
				if prep.Notes != "" {
					order.Items[i].Notes = prep.Notes
				}
			}
		}
	}

	order.Status = domain.MTDeliveryLoading
	order.PreparedBy = preparedBy

	if err := u.repo.Update(order); err != nil {
		return err
	}

	// Update Sales Order to LOADING and send FCM
	if order.SalesOrderID != nil {
		if err := u.db.Model(&domain.SalesOrder{}).Where("id = ?", *order.SalesOrderID).Update("status", "LOADING").Error; err == nil && u.fcmClient != nil {
			// Get employee id from sales order
			var so domain.SalesOrder
			if errSo := u.db.First(&so, "id = ?", *order.SalesOrderID).Error; errSo == nil {
				var employee domain.Employee
				if errEmp := u.db.First(&employee, "id = ?", so.EmployeeID).Error; errEmp == nil {
					title := "Barang MT Siap Dimuat"
					body := fmt.Sprintf("Barang untuk PO %s telah disiapkan oleh Gudang. Konfirmasi muatan dan mulai pengiriman.", so.SONumber)
					go func() {
						_ = u.fcmClient.SendPushNotification(employee.UserID, title, body)
					}()
				}
			}
		}
	}

	return nil
}

func (u *mtDeliveryUsecase) ShipDO(id uuid.UUID, vehicleID, driverID *uuid.UUID, notes string) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != domain.MTDeliveryLoading && order.Status != domain.MTDeliveryApproved {
		return errors.New("DO harus berstatus APPROVED atau LOADING untuk bisa dikirim")
	}
	now := time.Now()
	order.Status = domain.MTDeliveryShipped
	order.ShippedAt = &now
	order.VehicleID = vehicleID
	order.DriverID = driverID
	if notes != "" {
		order.Notes = notes
	}
	order.Items = nil
	return u.repo.Update(order)
}

func (u *mtDeliveryUsecase) MarkDelivered(id uuid.UUID) error {
	order, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != domain.MTDeliveryShipped {
		return errors.New("hanya DO berstatus SHIPPED yang bisa ditandai sudah terkirim")
	}
	now := time.Now()
	order.Status = domain.MTDeliveryDelivered
	order.DeliveredAt = &now
	order.Items = nil
	return u.repo.Update(order)
}
