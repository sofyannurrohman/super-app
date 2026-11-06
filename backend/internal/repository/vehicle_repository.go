package repository

import (
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type VehicleRepository interface {
	CreateWithComponents(vehicle *domain.Vehicle, components []domain.VehicleComponent) error
	GetAll(companyID *uuid.UUID, branchID *uuid.UUID) ([]domain.Vehicle, error)
	GetByID(id uuid.UUID) (*domain.Vehicle, error)
	UpdateWithComponents(vehicle *domain.Vehicle, components []domain.VehicleComponent) error
	Delete(id uuid.UUID) error
	GetLogs(vehicleID uuid.UUID) ([]domain.DeliveryBatch, error)
	GetActiveAssignment(vehicleID uuid.UUID) (*domain.ActiveDeliveryAssignment, error)
	RecordMaintenance(maintenance *domain.VehicleMaintenance) error
	AddMileage(vehicleID uuid.UUID, distance int) error
}

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &vehicleRepository{db}
}

func (r *vehicleRepository) CreateWithComponents(vehicle *domain.Vehicle, components []domain.VehicleComponent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(vehicle).Error; err != nil {
			return err
		}
		for i := range components {
			components[i].VehicleID = vehicle.ID
			if err := tx.Create(&components[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *vehicleRepository) GetAll(companyID *uuid.UUID, branchID *uuid.UUID) ([]domain.Vehicle, error) {
	var vehicles []domain.Vehicle
	db := r.db.Preload("Branch").Preload("Components")
	if companyID != nil && *companyID != uuid.Nil {
		db = db.Where("company_id = ?", *companyID)
	}
	if branchID != nil && *branchID != uuid.Nil {
		db = db.Where("branch_id = ?", *branchID)
	}
	err := db.Order("name ASC").Find(&vehicles).Error
	if err != nil {
		return nil, err
	}

	for i := range vehicles {
		assignment, _ := r.GetActiveAssignment(vehicles[i].ID)
		if assignment != nil {
			vehicles[i].ActiveAssignment = assignment
		}
	}

	return vehicles, nil
}

func (r *vehicleRepository) GetByID(id uuid.UUID) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.Preload("Branch").Preload("Components").First(&vehicle, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	assignment, _ := r.GetActiveAssignment(vehicle.ID)
	vehicle.ActiveAssignment = assignment
	return &vehicle, nil
}

func (r *vehicleRepository) UpdateWithComponents(vehicle *domain.Vehicle, components []domain.VehicleComponent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(vehicle).Error; err != nil {
			return err
		}
		// Sederhananya, untuk update komponen, kita hapus yang lama (yang tidak aktif atau semua) lalu recreate?
		// Karena ini komponen, mungkin kita update atau biarkan. Di sini kita hapus komponen yang lama jika ingin di-replace semua,
		// Atau lebih baik upsert. Untuk simpel: kita delete semua komponen kendaraan ini dan recreate (hati2 data maintenance hilang jika CASCADE)
		// Oh, karena foreign key ON DELETE CASCADE, kita tidak boleh delete.
		// Sebaiknya kita upsert.
		for _, comp := range components {
			comp.VehicleID = vehicle.ID
			if comp.ID == uuid.Nil {
				if err := tx.Create(&comp).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Save(&comp).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *vehicleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Vehicle{}, "id = ?", id).Error
}

func (r *vehicleRepository) GetLogs(vehicleID uuid.UUID) ([]domain.DeliveryBatch, error) {
	var logs []domain.DeliveryBatch
	err := r.db.Preload("Driver").
		Preload("Items.SalesTransaction.Store").
		Where("vehicle_id = ?", vehicleID).
		Order("created_at desc").
		Limit(50).
		Find(&logs).Error
	return logs, err
}

func (r *vehicleRepository) GetActiveAssignment(vehicleID uuid.UUID) (*domain.ActiveDeliveryAssignment, error) {
	// 1. Check DeliveryBatch (Sales Retail Surat Jalan)
	var activeBatch domain.DeliveryBatch
	err := r.db.Preload("Driver").
		Preload("DailyDriver").
		Preload("Helper").
		Preload("DailyHelper").
		Where("vehicle_id = ? AND status IN (?, ?, ?, ?, ?)",
			vehicleID,
			domain.DeliveryBatchWaitingApproval,
			domain.DeliveryBatchSupervisorApproved,
			domain.DeliveryBatchPicking,
			domain.DeliveryBatchOnDelivery,
			domain.DeliveryBatchWaitingCrosscheck,
		).
		Order("created_at desc").
		First(&activeBatch).Error

	if err == nil {
		doNo := "SJ-BATCH"
		if activeBatch.DeliveryOrderNo != nil && *activeBatch.DeliveryOrderNo != "" {
			doNo = *activeBatch.DeliveryOrderNo
		}

		driverName := "Belum Ditentukan"
		if activeBatch.Driver != nil {
			driverName = activeBatch.Driver.FirstName + " " + activeBatch.Driver.LastName
		} else if activeBatch.DailyDriver != nil {
			driverName = activeBatch.DailyDriver.Name + " (Harian)"
		}

		helperName := ""
		if activeBatch.Helper != nil {
			helperName = activeBatch.Helper.FirstName + " " + activeBatch.Helper.LastName
		} else if activeBatch.DailyHelper != nil {
			helperName = activeBatch.DailyHelper.Name + " (Harian)"
		}

		statusLabel := string(activeBatch.Status)
		switch activeBatch.Status {
		case domain.DeliveryBatchOnDelivery:
			statusLabel = "Sedang Dikirim"
		case domain.DeliveryBatchPicking:
			statusLabel = "Muat Barang Gudang"
		case domain.DeliveryBatchSupervisorApproved:
			statusLabel = "Disetujui SPV"
		case domain.DeliveryBatchWaitingApproval:
			statusLabel = "Menunggu SPV"
		case domain.DeliveryBatchWaitingCrosscheck:
			statusLabel = "Selesai Kirim (Setoran)"
		}

		var itemCount int64
		r.db.Table("delivery_items").Where("delivery_batch_id = ?", activeBatch.ID).Count(&itemCount)

		return &domain.ActiveDeliveryAssignment{
			DeliveryType:    domain.ActiveDeliveryTypeSalesBatch,
			TypeLabel:       "Pengiriman Sales Cabang",
			ReferenceID:     activeBatch.ID,
			DONumber:        doNo,
			Status:          string(activeBatch.Status),
			StatusLabel:     statusLabel,
			DriverName:      driverName,
			HelperName:      helperName,
			OriginName:      "Gudang Cabang",
			DestinationName: "Toko / Customer Pelanggan",
			TotalItems:      int(itemCount),
			StartedAt:       activeBatch.StartedAt,
			CreatedAt:       activeBatch.CreatedAt,
			RedirectPath:    "/sales/surat-jalan",
		}, nil
	}

	// 2. Check ProductTransfer (Pindah Gudang / Transfer Pabrik)
	var activeTransfer domain.ProductTransfer
	err = r.db.Preload("Driver").
		Preload("FromFactory").
		Preload("ToBranch").
		Where("vehicle_id = ? AND status IN (?, ?, ?, ?)",
			vehicleID,
			"APPROVED", "PREPARING", "SHIPPED", "ARRIVED",
		).
		Order("created_at desc").
		First(&activeTransfer).Error

	if err == nil {
		doNo := "DO-TRANSFER"
		if activeTransfer.DeliveryOrderNo != "" {
			doNo = activeTransfer.DeliveryOrderNo
		}

		driverName := "Belum Ditentukan"
		if activeTransfer.Driver != nil {
			driverName = activeTransfer.Driver.FirstName + " " + activeTransfer.Driver.LastName
		}

		originName := "Pabrik"
		if activeTransfer.FromFactory != nil {
			originName = activeTransfer.FromFactory.Name
		}

		destinationName := "Cabang"
		if activeTransfer.ToBranch != nil {
			destinationName = activeTransfer.ToBranch.Name
		}

		statusLabel := activeTransfer.Status
		switch activeTransfer.Status {
		case "SHIPPED":
			statusLabel = "Dalam Perjalanan"
		case "ARRIVED":
			statusLabel = "Tiba di Tujuan"
		case "PREPARING":
			statusLabel = "Persiapan Muatan"
		case "APPROVED":
			statusLabel = "Disetujui"
		}

		return &domain.ActiveDeliveryAssignment{
			DeliveryType:    domain.ActiveDeliveryTypeFactoryTransfer,
			TypeLabel:       "Pindah Gudang Pabrik",
			ReferenceID:     activeTransfer.ID,
			DONumber:        doNo,
			Status:          activeTransfer.Status,
			StatusLabel:     statusLabel,
			DriverName:      driverName,
			OriginName:      originName,
			DestinationName: destinationName,
			TotalItems:      activeTransfer.Quantity,
			CreatedAt:       activeTransfer.CreatedAt,
			RedirectPath:    "/factory/transfers",
		}, nil
	}

	// 3. Check MTDeliveryOrder (Modern Trade)
	var activeMT domain.MTDeliveryOrder
	err = r.db.Preload("Driver").
		Preload("KeyAccount").
		Preload("SourceFactory").
		Preload("SourceBranch").
		Where("vehicle_id = ? AND status IN (?, ?, ?)",
			vehicleID,
			"APPROVED", "PREPARING", "DISPATCHED",
		).
		Order("created_at desc").
		First(&activeMT).Error

	if err == nil {
		driverName := "Belum Ditentukan"
		if activeMT.Driver != nil {
			driverName = activeMT.Driver.FirstName + " " + activeMT.Driver.LastName
		}

		originName := "Pusat / Gudang"
		if activeMT.SourceFactory != nil {
			originName = activeMT.SourceFactory.Name
		} else if activeMT.SourceBranch != nil {
			originName = activeMT.SourceBranch.Name
		}

		destinationName := activeMT.DCDestination
		if activeMT.KeyAccount != nil {
			destinationName = activeMT.KeyAccount.Name + " (" + activeMT.DCDestination + ")"
		}

		statusLabel := string(activeMT.Status)
		if activeMT.Status == "DISPATCHED" {
			statusLabel = "Dalam Pengiriman"
		}

		return &domain.ActiveDeliveryAssignment{
			DeliveryType:    domain.ActiveDeliveryTypeMTDelivery,
			TypeLabel:       "Pengiriman Modern Trade",
			ReferenceID:     activeMT.ID,
			DONumber:        activeMT.DONumber,
			Status:          string(activeMT.Status),
			StatusLabel:     statusLabel,
			DriverName:      driverName,
			OriginName:      originName,
			DestinationName: destinationName,
			StartedAt:       activeMT.ShippedAt,
			CreatedAt:       activeMT.CreatedAt,
			RedirectPath:    "/sales/modern-trade/po",
		}, nil
	}

	// 4. Check FactoryReturn (Retur Pabrik)
	var activeReturn domain.FactoryReturn
	err = r.db.Preload("Driver").
		Preload("FromBranch").
		Where("vehicle_id = ? AND status IN (?, ?)",
			vehicleID,
			"SHIPPED", "ON_DELIVERY",
		).
		Order("created_at desc").
		First(&activeReturn).Error

	if err == nil {
		driverName := "Belum Ditentukan"
		if activeReturn.Driver != nil {
			driverName = activeReturn.Driver.FirstName + " " + activeReturn.Driver.LastName
		}

		originName := "Cabang"
		if activeReturn.FromBranch != nil {
			originName = activeReturn.FromBranch.Name
		}

		return &domain.ActiveDeliveryAssignment{
			DeliveryType:    domain.ActiveDeliveryTypeFactoryReturn,
			TypeLabel:       "Retur Barang Pabrik",
			ReferenceID:     activeReturn.ID,
			DONumber:        activeReturn.ReturnNo,
			Status:          activeReturn.Status,
			StatusLabel:     "Sedang Diantar ke Pabrik",
			DriverName:      driverName,
			OriginName:      originName,
			DestinationName: "Pabrik Utama",
			CreatedAt:       activeReturn.CreatedAt,
			RedirectPath:    "/factory/returns",
		}, nil
	}

	return nil, nil
}

func (r *vehicleRepository) RecordMaintenance(maintenance *domain.VehicleMaintenance) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(maintenance).Error; err != nil {
			return err
		}
		// Reset current_mileage of the component to 0
		return tx.Model(&domain.VehicleComponent{}).
			Where("id = ?", maintenance.VehicleComponentID).
			Update("current_mileage", 0).Error
	})
}

func (r *vehicleRepository) AddMileage(vehicleID uuid.UUID, distance int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update vehicle Spedometer
		if err := tx.Model(&domain.Vehicle{}).Where("id = ?", vehicleID).
			Update("mileage", gorm.Expr("mileage + ?", distance)).Error; err != nil {
			return err
		}

		// 2. Update active components current mileage
		if err := tx.Model(&domain.VehicleComponent{}).Where("vehicle_id = ? AND is_active = ?", vehicleID, true).
			Update("current_mileage", gorm.Expr("current_mileage + ?", distance)).Error; err != nil {
			return err
		}
		return nil
	})
}
