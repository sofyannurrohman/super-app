package usecase

import "github.com/google/uuid"

type DelegatedShipmentSummary struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	TotalPieces int       `json:"total_pieces"`
}
