package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type RawMaterialPOUsecase interface {
	// Purchase Request
	CreatePR(req CreateRawMaterialPRRequest, companyID uuid.UUID, userID *uuid.UUID, userName string) (*domain.RawMaterialPurchaseRequest, error)
	GetPRByID(id uuid.UUID) (*domain.RawMaterialPurchaseRequest, error)
	ListPRs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseRequest, error)
	UpdatePR(id uuid.UUID, req CreateRawMaterialPRRequest) (*domain.RawMaterialPurchaseRequest, error)
	RejectPR(id uuid.UUID, rejectedByID *uuid.UUID, rejectedName, reason string) (*domain.RawMaterialPurchaseRequest, error)
	CompletePR(id uuid.UUID, reviewedByID *uuid.UUID, reviewedName, notes string) (*domain.RawMaterialPurchaseRequest, error)
	DeletePR(id uuid.UUID) error

	// Purchase Order
	CreatePO(req CreateRawMaterialPORequest, companyID uuid.UUID, userID *uuid.UUID, userName string) (*domain.RawMaterialPurchaseOrder, error)
	GetPOByID(id uuid.UUID) (*domain.RawMaterialPurchaseOrder, error)
	ListPOs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseOrder, error)
	UpdatePO(id uuid.UUID, req CreateRawMaterialPORequest) (*domain.RawMaterialPurchaseOrder, error)
	ApprovePOBySPV(id uuid.UUID, spvUserID *uuid.UUID, spvName, spvPosition, notes string) (*domain.RawMaterialPurchaseOrder, error)
	ApprovePOByPurchasing(id uuid.UUID, purchUserID *uuid.UUID, purchName, purchPosition, notes string) (*domain.RawMaterialPurchaseOrder, error)
	UpdatePOProgress(id uuid.UUID, req UpdatePOProgressRequest, updatedByID *uuid.UUID, updatedByName string) (*domain.RawMaterialPurchaseOrder, error)
	UpdatePOPaymentStatus(id uuid.UUID, req UpdatePOPaymentStatusRequest) (*domain.RawMaterialPurchaseOrder, error)
	RejectPO(id uuid.UUID, rejectedByID *uuid.UUID, rejectedName, reason string) (*domain.RawMaterialPurchaseOrder, error)
	DeletePO(id uuid.UUID) error

	// Receipts
	CreateReceipt(req CreateRawMaterialReceiptRequest, companyID uuid.UUID, receivedByID *uuid.UUID, receivedByName string) (*domain.RawMaterialReceipt, error)
	GetReceiptByID(id uuid.UUID) (*domain.RawMaterialReceipt, error)
	ListReceipts(companyID, factoryID uuid.UUID) ([]domain.RawMaterialReceipt, error)
	PayReceipt(id uuid.UUID, amount float64) error
	ApproveReceiptPayment(id uuid.UUID) error
}

type CreateRawMaterialPRItemRequest struct {
	ProductID          *uuid.UUID `json:"product_id"`
	ItemName           string     `json:"item_name"`
	FactoryWarehouseID *uuid.UUID `json:"factory_warehouse_id"`
	RequestedQty       int        `json:"requested_qty" binding:"required,min=1"`
	Unit               string     `json:"unit"`
	Notes              string     `json:"notes"`
}

type CreateRawMaterialPRRequest struct {
	FactoryID    *uuid.UUID                       `json:"factory_id"`
	RequestDate  *time.Time                       `json:"request_date"`
	RequiredDate *time.Time                       `json:"required_date"`
	Notes        string                           `json:"notes"`
	Items        []CreateRawMaterialPRItemRequest `json:"items" binding:"required,min=1"`
}

type CreateRawMaterialPOItemRequest struct {
	ProductID          uuid.UUID  `json:"product_id" binding:"required"`
	FactoryWarehouseID *uuid.UUID `json:"factory_warehouse_id"`
	RequestedQty       int        `json:"requested_qty" binding:"required,min=1"`
	Unit               string     `json:"unit"`
	EstimatedUnitPrice float64    `json:"estimated_unit_price"`
	Notes              string     `json:"notes"`
}

type CreateRawMaterialPORequest struct {
	FactoryID         uuid.UUID                        `json:"factory_id" binding:"required"`
	SupplierID        uuid.UUID                        `json:"supplier_id" binding:"required"`
	PurchaseRequestID *uuid.UUID                       `json:"purchase_request_id"`
	OrderDate         *time.Time                       `json:"order_date"`
	EstimatedArrival  *time.Time                       `json:"estimated_arrival"`
	Notes             string                           `json:"notes"`
	Items             []CreateRawMaterialPOItemRequest `json:"items" binding:"required,min=1"`
}

type UpdatePOProgressRequest struct {
	ProgressStatus   string     `json:"progress_status" binding:"required"`
	ProgressNotes    string     `json:"progress_notes"`
	EstimatedArrival *time.Time `json:"estimated_arrival"`
}

type UpdatePOPaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status" binding:"required"`
	PaymentNotes  string `json:"payment_notes"`
}

type CreateRawMaterialReceiptItemRequest struct {
	POItemID           *uuid.UUID `json:"po_item_id"`
	ProductID          *uuid.UUID `json:"product_id"`
	TechnicianItemID   *uuid.UUID `json:"technician_item_id"`
	ItemType           string     `json:"item_type"` // RAW_MATERIAL, FINISHED_GOODS, TECHNICIAN_ITEM
	FactoryWarehouseID *uuid.UUID `json:"factory_warehouse_id"`
	BatchNo            string     `json:"batch_no"`
	ExpiryDate         *time.Time `json:"expiry_date"`
	ReceivedQty        int        `json:"received_qty" binding:"required,min=0"`
	RejectedQty        int        `json:"rejected_qty"`
	Unit               string     `json:"unit"`
	UnitPrice          float64    `json:"unit_price"`
	Notes              string     `json:"notes"`
}

type CreateRawMaterialReceiptRequest struct {
	FactoryID              uuid.UUID                             `json:"factory_id" binding:"required"`
	PurchaseOrderID        *uuid.UUID                            `json:"purchase_order_id"`
	SupplierID             *uuid.UUID                            `json:"supplier_id"`
	SupplierName           string                                `json:"supplier_name"`
	ReceiptType            string                                `json:"receipt_type"` // RAW_MATERIAL, TECHNICIAN_ITEM, MIXED
	SupplierDeliveryNoteNo string                                `json:"supplier_delivery_note_no"`
	SupplierInvoiceNo      string                                `json:"supplier_invoice_no"`
	InvoiceDate            *time.Time                            `json:"invoice_date"`
	PaymentDueDate         *time.Time                            `json:"payment_due_date"`
	DeliveryDate           *time.Time                            `json:"delivery_date"`
	ReceivedDate           *time.Time                            `json:"received_date"`
	ProofImageURL          string                                `json:"proof_image_url"`
	Notes                  string                                `json:"notes"`
	UnloadingCost          float64                               `json:"unloading_cost"`
	UnloadingNotes         string                                `json:"unloading_notes"`
	KasPabrikID            *uuid.UUID                            `json:"kas_pabrik_id"`
	Items                  []CreateRawMaterialReceiptItemRequest `json:"items" binding:"required,min=1"`
}

type rawMaterialPOUsecase struct {
	repo repository.RawMaterialPORepository
	db   *gorm.DB
}

func NewRawMaterialPOUsecase(repo repository.RawMaterialPORepository, db *gorm.DB) RawMaterialPOUsecase {
	return &rawMaterialPOUsecase{repo: repo, db: db}
}

func (u *rawMaterialPOUsecase) CreatePR(req CreateRawMaterialPRRequest, companyID uuid.UUID, userID *uuid.UUID, userName string) (*domain.RawMaterialPurchaseRequest, error) {
	prNumber := u.repo.GeneratePRNumber(companyID, req.FactoryID != nil)

	requestDate := time.Now()
	if req.RequestDate != nil {
		requestDate = *req.RequestDate
	}

	var items []domain.RawMaterialPurchaseRequestItem
	for _, itemReq := range req.Items {
		unit := strings.TrimSpace(itemReq.Unit)
		if itemReq.ProductID == nil && strings.TrimSpace(itemReq.ItemName) == "" {
			return nil, errors.New("nama barang wajib diisi")
		}
		if unit == "" && itemReq.ProductID != nil {
			var p domain.Product
			if err := u.db.Select("unit").First(&p, "id = ?", *itemReq.ProductID).Error; err == nil && p.Unit != "" {
				unit = p.Unit
			}
		}
		if unit == "" {
			unit = "PCS"
		}

		items = append(items, domain.RawMaterialPurchaseRequestItem{
			ProductID:          itemReq.ProductID,
			ItemName:           strings.TrimSpace(itemReq.ItemName),
			FactoryWarehouseID: itemReq.FactoryWarehouseID,
			RequestedQty:       itemReq.RequestedQty,
			Unit:               unit,
			Notes:              itemReq.Notes,
		})
	}

	pr := &domain.RawMaterialPurchaseRequest{
		CompanyID:     companyID,
		FactoryID:     req.FactoryID,
		PRNumber:      prNumber,
		RequestDate:   requestDate,
		RequiredDate:  req.RequiredDate,
		Status:        domain.RawMaterialPRPending,
		Notes:         req.Notes,
		CreatedByID:   userID,
		CreatedByName: userName,
		Items:         items,
	}

	if err := u.repo.CreatePR(pr); err != nil {
		return nil, err
	}

	return u.repo.GetPRByID(pr.ID)
}

func (u *rawMaterialPOUsecase) GetPRByID(id uuid.UUID) (*domain.RawMaterialPurchaseRequest, error) {
	return u.repo.GetPRByID(id)
}

func (u *rawMaterialPOUsecase) ListPRs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseRequest, error) {
	return u.repo.ListPRs(companyID, factoryID, status)
}

func (u *rawMaterialPOUsecase) UpdatePR(id uuid.UUID, req CreateRawMaterialPRRequest) (*domain.RawMaterialPurchaseRequest, error) {
	pr, err := u.repo.GetPRByID(id)
	if err != nil {
		return nil, err
	}

	if pr.Status != domain.RawMaterialPRPending && pr.Status != domain.RawMaterialPRRejected {
		return nil, errors.New("hanya Purchase Request berstatus Pending atau Ditolak yang dapat diedit")
	}

	requestDate := time.Now()
	if req.RequestDate != nil {
		requestDate = *req.RequestDate
	}

	var items []domain.RawMaterialPurchaseRequestItem
	for _, itemReq := range req.Items {
		unit := strings.TrimSpace(itemReq.Unit)
		if itemReq.ProductID == nil && strings.TrimSpace(itemReq.ItemName) == "" {
			return nil, errors.New("nama barang wajib diisi")
		}
		if unit == "" && itemReq.ProductID != nil {
			var p domain.Product
			if err := u.db.Select("unit").First(&p, "id = ?", *itemReq.ProductID).Error; err == nil && p.Unit != "" {
				unit = p.Unit
			}
		}
		if unit == "" {
			unit = "PCS"
		}

		items = append(items, domain.RawMaterialPurchaseRequestItem{
			PurchaseRequestID:  pr.ID,
			ProductID:          itemReq.ProductID,
			ItemName:           strings.TrimSpace(itemReq.ItemName),
			FactoryWarehouseID: itemReq.FactoryWarehouseID,
			RequestedQty:       itemReq.RequestedQty,
			Unit:               unit,
			Notes:              itemReq.Notes,
		})
	}

	pr.FactoryID = req.FactoryID
	pr.RequestDate = requestDate
	pr.RequiredDate = req.RequiredDate
	pr.Notes = req.Notes
	pr.Status = domain.RawMaterialPRPending
	pr.Items = items

	if err := u.repo.UpdatePR(pr); err != nil {
		return nil, err
	}

	return u.repo.GetPRByID(pr.ID)
}

func (u *rawMaterialPOUsecase) RejectPR(id uuid.UUID, rejectedByID *uuid.UUID, rejectedName, reason string) (*domain.RawMaterialPurchaseRequest, error) {
	pr, err := u.repo.GetPRByID(id)
	if err != nil {
		return nil, err
	}

	if pr.Status != domain.RawMaterialPRPending {
		return nil, errors.New("hanya Purchase Request yang berstatus Menunggu Review yang dapat ditolak")
	}

	if err := u.repo.UpdatePRStatus(id, domain.RawMaterialPRRejected, nil, rejectedByID, rejectedName, "", reason); err != nil {
		return nil, err
	}

	return u.repo.GetPRByID(id)
}

func (u *rawMaterialPOUsecase) CompletePR(id uuid.UUID, reviewedByID *uuid.UUID, reviewedName, notes string) (*domain.RawMaterialPurchaseRequest, error) {
	pr, err := u.repo.GetPRByID(id)
	if err != nil {
		return nil, err
	}
	if pr.Status != domain.RawMaterialPRPending {
		return nil, errors.New("hanya Purchase Request yang menunggu tindak lanjut dapat diselesaikan")
	}
	if err := u.repo.UpdatePRStatus(id, domain.RawMaterialPRApproved, nil, reviewedByID, reviewedName, strings.TrimSpace(notes), ""); err != nil {
		return nil, err
	}
	return u.repo.GetPRByID(id)
}

func (u *rawMaterialPOUsecase) DeletePR(id uuid.UUID) error {
	pr, err := u.repo.GetPRByID(id)
	if err != nil {
		return err
	}
	if pr.Status == domain.RawMaterialPRApproved {
		return errors.New("Purchase Request yang sudah ditindaklanjuti tidak dapat dihapus")
	}
	return u.repo.DeletePR(id)
}

func (u *rawMaterialPOUsecase) CreatePO(req CreateRawMaterialPORequest, companyID uuid.UUID, userID *uuid.UUID, userName string) (*domain.RawMaterialPurchaseOrder, error) {
	poNumber := u.repo.GeneratePONumber(companyID)

	orderDate := time.Now()
	if req.OrderDate != nil {
		orderDate = *req.OrderDate
	}

	var totalEstimated float64
	var items []domain.RawMaterialPurchaseOrderItem

	for _, itemReq := range req.Items {
		unit := strings.TrimSpace(itemReq.Unit)
		if unit == "" {
			var p domain.Product
			if err := u.db.Select("unit").First(&p, "id = ?", itemReq.ProductID).Error; err == nil && p.Unit != "" {
				unit = p.Unit
			} else {
				unit = "PCS"
			}
		}
		subtotal := float64(itemReq.RequestedQty) * itemReq.EstimatedUnitPrice
		totalEstimated += subtotal

		items = append(items, domain.RawMaterialPurchaseOrderItem{
			ProductID:          itemReq.ProductID,
			FactoryWarehouseID: itemReq.FactoryWarehouseID,
			RequestedQty:       itemReq.RequestedQty,
			ReceivedQty:        0,
			Unit:               unit,
			EstimatedUnitPrice: itemReq.EstimatedUnitPrice,
			Subtotal:           subtotal,
			Notes:              itemReq.Notes,
		})
	}

	po := &domain.RawMaterialPurchaseOrder{
		CompanyID:            companyID,
		FactoryID:            req.FactoryID,
		SupplierID:           req.SupplierID,
		PurchaseRequestID:    req.PurchaseRequestID,
		PONumber:             poNumber,
		OrderDate:            orderDate,
		EstimatedArrival:     req.EstimatedArrival,
		Status:               domain.RawMaterialPOPendingSPV,
		ProgressStatus:       "PENDING_APPROVAL",
		PaymentStatus:        "UNPAID",
		TotalEstimatedAmount: totalEstimated,
		Notes:                req.Notes,
		CreatedByID:          userID,
		CreatedByName:        userName,
		Items:                items,
	}

	if err := u.repo.CreatePO(po); err != nil {
		return nil, err
	}

	// If linked to a PR, mark PR as processed
	if req.PurchaseRequestID != nil && *req.PurchaseRequestID != uuid.Nil {
		_ = u.repo.UpdatePRStatus(
			*req.PurchaseRequestID,
			domain.RawMaterialPRApproved,
			&po.ID,
			userID,
			userName,
			"Diproses menjadi Purchase Order "+po.PONumber,
			"",
		)
	}

	return u.repo.GetPOByID(po.ID)
}

func (u *rawMaterialPOUsecase) GetPOByID(id uuid.UUID) (*domain.RawMaterialPurchaseOrder, error) {
	return u.repo.GetPOByID(id)
}

func (u *rawMaterialPOUsecase) ListPOs(companyID, factoryID uuid.UUID, status string) ([]domain.RawMaterialPurchaseOrder, error) {
	return u.repo.ListPOs(companyID, factoryID, status)
}

func (u *rawMaterialPOUsecase) UpdatePO(id uuid.UUID, req CreateRawMaterialPORequest) (*domain.RawMaterialPurchaseOrder, error) {
	po, err := u.repo.GetPOByID(id)
	if err != nil {
		return nil, err
	}

	if po.Status != domain.RawMaterialPODraft && po.Status != domain.RawMaterialPOPendingSPV && po.Status != domain.RawMaterialPORejected {
		return nil, errors.New("hanya PO berstatus Draft, Menunggu SPV, atau Ditolak yang dapat diedit")
	}

	orderDate := time.Now()
	if req.OrderDate != nil {
		orderDate = *req.OrderDate
	}

	var totalEstimated float64
	var items []domain.RawMaterialPurchaseOrderItem

	for _, itemReq := range req.Items {
		unit := strings.TrimSpace(itemReq.Unit)
		if unit == "" {
			var p domain.Product
			if err := u.db.Select("unit").First(&p, "id = ?", itemReq.ProductID).Error; err == nil && p.Unit != "" {
				unit = p.Unit
			} else {
				unit = "PCS"
			}
		}
		subtotal := float64(itemReq.RequestedQty) * itemReq.EstimatedUnitPrice
		totalEstimated += subtotal

		items = append(items, domain.RawMaterialPurchaseOrderItem{
			PurchaseOrderID:    po.ID,
			ProductID:          itemReq.ProductID,
			FactoryWarehouseID: itemReq.FactoryWarehouseID,
			RequestedQty:       itemReq.RequestedQty,
			ReceivedQty:        0,
			Unit:               unit,
			EstimatedUnitPrice: itemReq.EstimatedUnitPrice,
			Subtotal:           subtotal,
			Notes:              itemReq.Notes,
		})
	}

	po.SupplierID = req.SupplierID
	po.FactoryID = req.FactoryID
	po.OrderDate = orderDate
	po.EstimatedArrival = req.EstimatedArrival
	po.Notes = req.Notes
	po.TotalEstimatedAmount = totalEstimated
	po.Status = domain.RawMaterialPOPendingSPV
	po.Items = items

	if err := u.repo.UpdatePO(po); err != nil {
		return nil, err
	}

	return u.repo.GetPOByID(po.ID)
}

func (u *rawMaterialPOUsecase) ApprovePOBySPV(id uuid.UUID, spvUserID *uuid.UUID, spvName, spvPosition, notes string) (*domain.RawMaterialPurchaseOrder, error) {
	po, err := u.repo.GetPOByID(id)
	if err != nil {
		return nil, err
	}

	if po.Status != domain.RawMaterialPOPendingSPV {
		return nil, errors.New("PO tidak dalam status menunggu persetujuan SPV")
	}

	now := time.Now()
	po.Status = domain.RawMaterialPOPendingPurchasing
	po.SPVApprovedByID = spvUserID
	po.SPVApprovedByName = spvName
	po.SPVApprovedByPosition = spvPosition
	po.SPVApprovedAt = &now
	po.SPVNotes = notes

	if err := u.repo.UpdatePO(po); err != nil {
		return nil, err
	}

	return u.repo.GetPOByID(po.ID)
}

func (u *rawMaterialPOUsecase) ApprovePOByPurchasing(id uuid.UUID, purchUserID *uuid.UUID, purchName, purchPosition, notes string) (*domain.RawMaterialPurchaseOrder, error) {
	po, err := u.repo.GetPOByID(id)
	if err != nil {
		return nil, err
	}

	if po.Status != domain.RawMaterialPOPendingPurchasing {
		return nil, errors.New("PO tidak dalam status menunggu persetujuan Purchasing")
	}

	now := time.Now()
	po.Status = domain.RawMaterialPOApproved
	po.ProgressStatus = "PO_SENT_TO_SUPPLIER"
	po.PurchasingApprovedByID = purchUserID
	po.PurchasingApprovedByName = purchName
	po.PurchasingApprovedByPosition = purchPosition
	po.PurchasingApprovedAt = &now
	po.PurchasingNotes = notes
	po.ProgressUpdatedByID = purchUserID
	po.ProgressUpdatedByName = purchName
	po.ProgressUpdatedAt = &now

	if err := u.repo.UpdatePO(po); err != nil {
		return nil, err
	}

	return u.repo.GetPOByID(po.ID)
}

func (u *rawMaterialPOUsecase) UpdatePOProgress(id uuid.UUID, req UpdatePOProgressRequest, updatedByID *uuid.UUID, updatedByName string) (*domain.RawMaterialPurchaseOrder, error) {
	po, err := u.repo.GetPOByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	po.ProgressStatus = req.ProgressStatus
	if req.ProgressNotes != "" {
		po.ProgressNotes = req.ProgressNotes
	}
	if req.EstimatedArrival != nil {
		po.EstimatedArrival = req.EstimatedArrival
	}
	po.ProgressUpdatedByID = updatedByID
	po.ProgressUpdatedByName = updatedByName
	po.ProgressUpdatedAt = &now

	if err := u.repo.UpdatePO(po); err != nil {
		return nil, err
	}

	return u.repo.GetPOByID(po.ID)
}

func (u *rawMaterialPOUsecase) UpdatePOPaymentStatus(id uuid.UUID, req UpdatePOPaymentStatusRequest) (*domain.RawMaterialPurchaseOrder, error) {
	po, err := u.repo.GetPOByID(id)
	if err != nil {
		return nil, err
	}

	po.PaymentStatus = req.PaymentStatus
	if req.PaymentNotes != "" {
		po.PaymentNotes = req.PaymentNotes
	}

	if err := u.repo.UpdatePO(po); err != nil {
		return nil, err
	}

	return u.repo.GetPOByID(po.ID)
}

func (u *rawMaterialPOUsecase) RejectPO(id uuid.UUID, rejectedByID *uuid.UUID, rejectedName, reason string) (*domain.RawMaterialPurchaseOrder, error) {
	po, err := u.repo.GetPOByID(id)
	if err != nil {
		return nil, err
	}

	if po.Status == domain.RawMaterialPOReceived || po.Status == domain.RawMaterialPOPartialReceived {
		return nil, errors.New("PO yang sudah memiliki penerimaan barang tidak dapat ditolak")
	}

	now := time.Now()
	po.Status = domain.RawMaterialPORejected
	po.ProgressStatus = "REJECTED"
	po.RejectedByID = rejectedByID
	po.RejectedByName = rejectedName
	po.RejectedAt = &now
	po.RejectionReason = reason

	if err := u.repo.UpdatePO(po); err != nil {
		return nil, err
	}

	return u.repo.GetPOByID(po.ID)
}

func (u *rawMaterialPOUsecase) DeletePO(id uuid.UUID) error {
	_, err := u.repo.GetPOByID(id)
	if err != nil {
		return err
	}

	return u.repo.DeletePO(id)
}

func (u *rawMaterialPOUsecase) CreateReceipt(req CreateRawMaterialReceiptRequest, companyID uuid.UUID, receivedByID *uuid.UUID, receivedByName string) (*domain.RawMaterialReceipt, error) {
	receiptNo := u.repo.GenerateReceiptNumber(companyID)

	receivedDate := time.Now()
	if req.ReceivedDate != nil {
		receivedDate = *req.ReceivedDate
	}

	var items []domain.RawMaterialReceiptItem
	var invoiceTotal float64
	var po *domain.RawMaterialPurchaseOrder
	var err error

	if req.PurchaseOrderID != nil && *req.PurchaseOrderID != uuid.Nil {
		po, err = u.repo.GetPOByID(*req.PurchaseOrderID)
		if err != nil {
			return nil, errors.New("Purchase Order tidak ditemukan")
		}
		if po.Status != domain.RawMaterialPOApproved && po.Status != domain.RawMaterialPOPartialReceived {
			return nil, errors.New("penerimaan hanya bisa dicatat untuk PO yang sudah disetujui (APPROVED / PARTIAL_RECEIVED)")
		}
		if req.SupplierID != nil && *req.SupplierID != uuid.Nil && *req.SupplierID != po.SupplierID {
			return nil, errors.New("supplier penerimaan harus sama dengan supplier pada PO")
		}
	} else {
		// Receipt without PO is allowed, but must be explicitly attributable.
		if (req.SupplierID == nil || *req.SupplierID == uuid.Nil) && strings.TrimSpace(req.SupplierName) == "" {
			return nil, errors.New("supplier wajib dipilih untuk penerimaan tanpa PO")
		}
	}

	for _, itemReq := range req.Items {
		unit := strings.TrimSpace(itemReq.Unit)
		if unit == "" {
			if itemReq.ProductID != nil && *itemReq.ProductID != uuid.Nil {
				var p domain.Product
				if err := u.db.Select("unit").First(&p, "id = ?", *itemReq.ProductID).Error; err == nil && p.Unit != "" {
					unit = p.Unit
				} else {
					unit = "PCS"
				}
			} else if itemReq.TechnicianItemID != nil && *itemReq.TechnicianItemID != uuid.Nil {
				var t domain.TechnicianItem
				if err := u.db.Select("unit").First(&t, "id = ?", *itemReq.TechnicianItemID).Error; err == nil && t.Unit != "" {
					unit = t.Unit
				} else {
					unit = "PCS"
				}
			} else {
				unit = "PCS"
			}
		}
		batchNo := itemReq.BatchNo
		if batchNo == "" {
			if itemReq.ProductID != nil && *itemReq.ProductID != uuid.Nil {
				batchNo = generateInventoryBatch("RM", *itemReq.ProductID, receivedDate)
			} else if itemReq.TechnicianItemID != nil && *itemReq.TechnicianItemID != uuid.Nil {
				batchNo = generateInventoryBatch("TECH", *itemReq.TechnicianItemID, receivedDate)
			} else {
				return nil, errors.New("produk atau item teknisi wajib dipilih")
			}
		}

		itemType := "RAW_MATERIAL"
		if itemReq.ItemType == "TECHNICIAN_ITEM" || (itemReq.TechnicianItemID != nil && *itemReq.TechnicianItemID != uuid.Nil) {
			itemType = "TECHNICIAN_ITEM"
		} else if itemReq.ItemType == "FINISHED_GOODS" {
			itemType = "FINISHED_GOODS"
		}

		// Prevent finished goods from being silently recorded in a raw-material
		// warehouse (and the reverse). The stock table is shared, so the
		// warehouse type is the authoritative classification at receipt time.
		if itemReq.FactoryWarehouseID != nil && *itemReq.FactoryWarehouseID != uuid.Nil {
			var warehouse domain.FactoryWarehouse
			if err := u.db.Select("type").First(&warehouse, "id = ?", *itemReq.FactoryWarehouseID).Error; err != nil {
				return nil, errors.New("gudang tujuan penerimaan tidak ditemukan")
			}
			warehouseType := strings.ToUpper(strings.TrimSpace(warehouse.Type))
			if itemType == "RAW_MATERIAL" && warehouseType == "FINISHED_GOODS" {
				return nil, errors.New("bahan baku tidak boleh masuk ke gudang barang jadi")
			}
			if itemType == "FINISHED_GOODS" && warehouseType == "RAW_MATERIAL" {
				return nil, errors.New("barang jadi tidak boleh masuk ke gudang bahan baku")
			}
		}

		receiptItem := domain.RawMaterialReceiptItem{
			POItemID:           itemReq.POItemID,
			ProductID:          itemReq.ProductID,
			TechnicianItemID:   itemReq.TechnicianItemID,
			ItemType:           itemType,
			FactoryWarehouseID: itemReq.FactoryWarehouseID,
			BatchNo:            batchNo,
			ExpiryDate:         itemReq.ExpiryDate,
			ReceivedQty:        itemReq.ReceivedQty,
			RejectedQty:        itemReq.RejectedQty,
			Unit:               unit,
			UnitPrice:          itemReq.UnitPrice,
			Notes:              itemReq.Notes,
		}
		items = append(items, receiptItem)
		invoiceTotal += float64(itemReq.ReceivedQty) * itemReq.UnitPrice
	}

	receiptType := req.ReceiptType
	if receiptType == "" {
		hasRaw := false
		hasFinished := false
		hasTech := false
		for _, it := range items {
			if it.ItemType == "TECHNICIAN_ITEM" {
				hasTech = true
			} else if it.ItemType == "FINISHED_GOODS" {
				hasFinished = true
			} else {
				hasRaw = true
			}
		}
		if (hasRaw && hasTech) || (hasRaw && hasFinished) || (hasTech && hasFinished) {
			receiptType = "MIXED"
		} else if hasFinished {
			receiptType = "FINISHED_GOODS"
		} else if hasTech {
			receiptType = "TECHNICIAN_ITEM"
		} else {
			receiptType = "RAW_MATERIAL"
		}
	}

	receipt := &domain.RawMaterialReceipt{
		CompanyID:              companyID,
		FactoryID:              req.FactoryID,
		PurchaseOrderID:        req.PurchaseOrderID,
		SupplierID:             req.SupplierID,
		SupplierName:           req.SupplierName,
		ReceiptType:            receiptType,
		ReceiptNo:              receiptNo,
		SupplierDeliveryNoteNo: req.SupplierDeliveryNoteNo,
		SupplierInvoiceNo:      req.SupplierInvoiceNo,
		InvoiceDate:            req.InvoiceDate,
		PaymentDueDate:         req.PaymentDueDate,
		DeliveryDate:           req.DeliveryDate,
		ReceivedDate:           receivedDate,
		ReceivedByID:           receivedByID,
		ReceivedByName:         receivedByName,
		ProofImageURL:          req.ProofImageURL,
		Notes:                  req.Notes,
		UnloadingCost:          req.UnloadingCost,
		UnloadingNotes:         req.UnloadingNotes,
		KasPabrikID:            req.KasPabrikID,
		Items:                  items,
		InvoiceTotal:           invoiceTotal,
		PaymentStatus:          "UNPAID",
	}

	if err := u.repo.CreateReceipt(receipt); err != nil {
		return nil, err
	}

	// Update Factory Stock / Technician Item Stock & Log for each item
	for _, item := range receipt.Items {
		if item.ReceivedQty > 0 {
			if item.ItemType == "TECHNICIAN_ITEM" && item.TechnicianItemID != nil && *item.TechnicianItemID != uuid.Nil {
				// Update Technician Item Stock
				var techItem domain.TechnicianItem
				if err := u.db.First(&techItem, "id = ?", *item.TechnicianItemID).Error; err == nil {
					newQty := techItem.Quantity + item.ReceivedQty
					_ = u.db.Model(&techItem).Update("quantity", newQty).Error

					supplierLabel := receipt.SupplierName
					if supplierLabel == "" && receipt.Supplier != nil {
						supplierLabel = receipt.Supplier.Name
					}
					if supplierLabel == "" {
						supplierLabel = "Supplier"
					}

					notes := item.Notes
					if notes == "" {
						notes = fmt.Sprintf("Penerimaan dari %s (SJ: %s)", supplierLabel, receipt.SupplierDeliveryNoteNo)
					}

					techLog := domain.TechnicianStockLog{
						CompanyID:        companyID,
						FactoryID:        receipt.FactoryID,
						TechnicianItemID: *item.TechnicianItemID,
						Type:             "IN",
						Source:           "SUPPLIER_RECEIPT",
						Quantity:         item.ReceivedQty,
						CurrentStock:     newQty,
						TechnicianName:   receivedByName,
						ReferenceNo:      receipt.SupplierDeliveryNoteNo,
						Notes:            notes,
						CreatedAt:        receivedDate,
					}
					_ = u.db.Create(&techLog).Error
				}
			} else if item.ProductID != nil && *item.ProductID != uuid.Nil {
				// Raw Material Stock
				if err := u.repo.UpsertRawMaterialStock(receipt.FactoryID, *item.ProductID, item.FactoryWarehouseID, item.BatchNo, item.ExpiryDate, item.ReceivedQty); err != nil {
					return nil, err
				}

				// Record in Factory Inventory Log
				log := domain.FactoryInventoryLog{
					FactoryID: receipt.FactoryID,
					ProductID: *item.ProductID,
					Type:      "IN",
					Source:    "RAW_MATERIAL_RECEIPT",
					BatchNo:   item.BatchNo,
					Quantity:  item.ReceivedQty,
					CreatedAt: receivedDate,
				}
				_ = u.repo.CreateInventoryLog(&log)
			}
		}
	}

	// If linked to PO, update PO items received_qty and PO status
	if po != nil {
		for _, rItem := range receipt.Items {
			for i := range po.Items {
				if (rItem.POItemID != nil && po.Items[i].ID == *rItem.POItemID) ||
					(rItem.POItemID == nil && rItem.ProductID != nil && po.Items[i].ProductID == *rItem.ProductID) {
					po.Items[i].ReceivedQty += rItem.ReceivedQty
				}
			}
		}

		allReceived := true
		someReceived := false
		for _, pItem := range po.Items {
			if pItem.ReceivedQty < pItem.RequestedQty {
				allReceived = false
			}
			if pItem.ReceivedQty > 0 {
				someReceived = true
			}
		}

		if allReceived {
			po.Status = domain.RawMaterialPOReceived
			po.ProgressStatus = "DELIVERED"
		} else if someReceived {
			po.Status = domain.RawMaterialPOPartialReceived
			po.ProgressStatus = "PARTIAL_DELIVERED"
		}

		_ = u.repo.UpdatePO(po)
	}

	// Unloading Cost Posting to Kas Pabrik (Mbak Linda)
	if req.UnloadingCost > 0 && req.KasPabrikID != nil && *req.KasPabrikID != uuid.Nil && u.db != nil {
		var kas domain.KasPabrik
		if err := u.db.First(&kas, "id = ?", *req.KasPabrikID).Error; err == nil {
			balanceBefore := kas.Balance
			balanceAfter := kas.Balance - req.UnloadingCost
			kas.Balance = balanceAfter
			_ = u.db.Save(&kas).Error

			desc := fmt.Sprintf("Ongkos Bongkar Surat Jalan: %s (No: %s)", receipt.SupplierDeliveryNoteNo, receipt.ReceiptNo)
			if req.UnloadingNotes != "" {
				desc += fmt.Sprintf(" - %s", req.UnloadingNotes)
			}

			txModel := domain.KasTransaction{
				KasPabrikID:     *req.KasPabrikID,
				TransactionType: "OUT",
				Amount:          req.UnloadingCost,
				BalanceBefore:   balanceBefore,
				BalanceAfter:    balanceAfter,
				ExpenseCategory: "BIAYA_BONGKAR_MUAT",
				Description:     desc,
				TransactionDate: receivedDate,
				CreatedBy:       receivedByID,
			}
			if err := u.db.Create(&txModel).Error; err == nil {
				receipt.KasTransactionID = &txModel.ID
				_ = u.db.Model(&domain.RawMaterialReceipt{}).Where("id = ?", receipt.ID).Update("kas_transaction_id", txModel.ID).Error
			}
		}
	}

	return u.repo.GetReceiptByID(receipt.ID)
}

func (u *rawMaterialPOUsecase) GetReceiptByID(id uuid.UUID) (*domain.RawMaterialReceipt, error) {
	return u.repo.GetReceiptByID(id)
}

func (u *rawMaterialPOUsecase) ListReceipts(companyID, factoryID uuid.UUID) ([]domain.RawMaterialReceipt, error) {
	return u.repo.ListReceipts(companyID, factoryID)
}

func (u *rawMaterialPOUsecase) PayReceipt(id uuid.UUID, amount float64) error {
	if amount <= 0 {
		return errors.New("jumlah pembayaran harus lebih dari nol")
	}
	return u.db.Transaction(func(tx *gorm.DB) error {
		var receipt domain.RawMaterialReceipt
		if err := tx.First(&receipt, "id = ?", id).Error; err != nil {
			return err
		}
		if receipt.InvoiceTotal <= 0 {
			return errors.New("total faktur belum tersedia")
		}
		if receipt.PaidAmount+amount > receipt.InvoiceTotal {
			return errors.New("pembayaran melebihi sisa hutang")
		}
		if receipt.PendingPaymentAmount > 0 {
			return errors.New("sudah ada pembayaran yang menunggu approval cashier")
		}
		receipt.PendingPaymentAmount = amount
		receipt.PaymentApprovalStatus = "PENDING_APPROVAL"
		return tx.Save(&receipt).Error
	})
}

func (u *rawMaterialPOUsecase) ApproveReceiptPayment(id uuid.UUID) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var receipt domain.RawMaterialReceipt
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&receipt, "id = ?", id).Error; err != nil {
			return err
		}
		amount := receipt.PendingPaymentAmount
		if receipt.PaymentApprovalStatus != "PENDING_APPROVAL" || amount <= 0 {
			return errors.New("tidak ada pembayaran pending")
		}
		if receipt.KasPabrikID == nil || *receipt.KasPabrikID == uuid.Nil {
			return errors.New("kas pabrik sumber pembayaran wajib dipilih")
		}
		var kas domain.KasPabrik
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", *receipt.KasPabrikID).Error; err != nil {
			return err
		}
		if kas.Balance < amount {
			return errors.New("saldo kas pabrik tidak mencukupi")
		}
		before := kas.Balance
		kas.Balance -= amount
		if err := tx.Save(&kas).Error; err != nil {
			return err
		}
		kasTx := &domain.KasTransaction{KasPabrikID: kas.ID, TransactionType: "OUT", Amount: amount, BalanceBefore: before, BalanceAfter: kas.Balance, ExpenseCategory: "SUPPLIER_PAYMENT", Description: fmt.Sprintf("Pembayaran supplier - faktur %s", receipt.SupplierInvoiceNo), TransactionDate: time.Now()}
		if err := tx.Create(kasTx).Error; err != nil {
			return err
		}
		receipt.PaidAmount += amount
		receipt.KasTransactionID = &kasTx.ID
		receipt.PendingPaymentAmount = 0
		receipt.PaymentApprovalStatus = "APPROVED"
		receipt.PaymentStatus = "PARTIAL"
		if receipt.PaidAmount >= receipt.InvoiceTotal {
			receipt.PaymentStatus = "PAID"
		}
		return tx.Save(&receipt).Error
	})
}
