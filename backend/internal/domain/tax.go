package domain

import (
	"github.com/google/uuid"
)

type SetTaxInvoiceRequest struct {
	SalesOrderID     uuid.UUID `json:"sales_order_id" binding:"required"`
	TaxInvoiceNumber string    `json:"tax_invoice_number" binding:"required"`
	TaxInvoiceDate   string    `json:"tax_invoice_date" binding:"required"`
	TaxRate          float64   `json:"tax_rate" binding:"required"`
}

type ExportEfakturRequest struct {
	CompanyID string `form:"company_id" binding:"required"`
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
}
