package usecase

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type TaxUsecase interface {
	SetTaxInvoice(req domain.SetTaxInvoiceRequest) error
	GenerateEfakturCSV(companyID uuid.UUID, startDate, endDate time.Time) ([]byte, error)
}

type taxUsecase struct {
	db *gorm.DB
}

func NewTaxUsecase(db *gorm.DB) TaxUsecase {
	return &taxUsecase{db: db}
}

func (u *taxUsecase) SetTaxInvoice(req domain.SetTaxInvoiceRequest) error {
	var so domain.SalesOrder
	if err := u.db.First(&so, "id = ?", req.SalesOrderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("sales order not found")
		}
		return err
	}

	// Hitung DPP dan PPN
	// Asumsi TotalAmount sudah termasuk PPN (Include). Jika exclude, perhitungannya beda.
	// Kita asumsikan TotalAmount adalah nilai kotor (DPP + PPN).
	// DPP = TotalAmount / (1 + (TaxRate / 100))
	taxRateDecimal := req.TaxRate / 100.0
	dpp := so.TotalAmount / (1.0 + taxRateDecimal)
	ppn := so.TotalAmount - dpp

	parsedDate, err := time.Parse("2006-01-02", req.TaxInvoiceDate)
	if err != nil {
		return errors.New("invalid date format, expected YYYY-MM-DD")
	}

	so.IsTaxable = true
	so.TaxInvoiceNumber = &req.TaxInvoiceNumber
	so.TaxInvoiceDate = &parsedDate
	so.TaxRate = req.TaxRate
	so.DPPAmount = dpp
	so.PPNAmount = ppn

	if err := u.db.Save(&so).Error; err != nil {
		return fmt.Errorf("failed to save tax invoice details: %w", err)
	}

	return nil
}

func (u *taxUsecase) GenerateEfakturCSV(companyID uuid.UUID, startDate, endDate time.Time) ([]byte, error) {
	var orders []domain.SalesOrder
	// Preload Store dan Store.Company
	err := u.db.Preload("Store").
		Where("company_id = ? AND is_taxable = ? AND tax_invoice_date BETWEEN ? AND ?", companyID, true, startDate, endDate).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	// Prepare CSV writing
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	// e-Faktur usually uses comma delimiter
	writer.Comma = ','

	// Write FK and OF rows
	for _, so := range orders {
		// FK (Faktur Keluaran) Row
		var npwp, nik, name, address string
		if so.Store.NPWP != nil {
			npwp = *so.Store.NPWP
		} else {
			npwp = "000000000000000" // NPWP 000 if not available
		}
		if so.Store.NIK != nil {
			nik = *so.Store.NIK
		}
		if so.Store.TaxName != nil {
			name = *so.Store.TaxName
		} else {
			name = so.Store.Name
		}
		if so.Store.Address != "" {
			address = so.Store.Address
		} else {
			address = "Alamat Lengkap"
		}

		taxDateStr := so.TaxInvoiceDate.Format("02/01/2006")
		masaPajak := so.TaxInvoiceDate.Format("01")
		tahunPajak := so.TaxInvoiceDate.Format("2006")

		nsfp := ""
		if so.TaxInvoiceNumber != nil {
			nsfp = *so.TaxInvoiceNumber
		}

		// DJP e-Faktur format FK
		fkRow := []string{
			"FK",                              // Header type
			"01",                              // KD_JENIS_TRANSAKSI
			"0",                               // FG_PENGGANTI
			nsfp,                              // NOMOR_FAKTUR
			masaPajak,                         // MASA_PAJAK
			tahunPajak,                        // TAHUN_PAJAK
			taxDateStr,                        // TANGGAL_FAKTUR
			npwp,                              // NPWP
			name,                              // NAMA
			address,                           // ALAMAT_LENGKAP
			fmt.Sprintf("%.0f", so.DPPAmount), // JUMLAH_DPP
			fmt.Sprintf("%.0f", so.PPNAmount), // JUMLAH_PPN
			"0",                               // JUMLAH_PPNBM
			"",                                // ID_KETERANGAN_TAMBAHAN
			"0",                               // FG_UANG_MUKA
			"0",                               // UANG_MUKA_DPP
			"0",                               // UANG_MUKA_PPN
			"0",                               // UANG_MUKA_PPNBM
			so.SONumber,                       // REFERENSI
			nik,                               // NIK (sometimes at the end depending on DJP version, or not needed. Usually NIK is only for NPWP 000. Let's keep it in NIK column if present)
		}

		// Note: The standard 2024 DJP format for FK actually has 19 columns if including NIK at the end,
		// but let's just write what we have. It usually throws warnings if mismatched, we'll try to stick to standard.
		if err := writer.Write(fkRow); err != nil {
			return nil, err
		}

		itemDpp := so.DPPAmount
		itemPpn := so.PPNAmount
		itemPrice := so.TotalAmount

		ofRow := []string{
			"OF",
			"BRG-01",                          // KODE_OBJEK
			"Penjualan Barang " + so.SONumber, // NAMA
			fmt.Sprintf("%.0f", itemPrice),    // HARGA_SATUAN
			"1",                               // JUMLAH_BARANG
			fmt.Sprintf("%.0f", itemPrice),    // HARGA_TOTAL
			"0",                               // DISKON
			fmt.Sprintf("%.0f", itemDpp),      // DPP
			fmt.Sprintf("%.0f", itemPpn),      // PPN
			"0",                               // TARIF_PPNBM
			"0",                               // PPNBM
		}
		if err := writer.Write(ofRow); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
