package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// generateInventoryBatch creates a traceable batch number for new inventory
// transactions. Existing DEFAULT batches remain valid for legacy stock.
func generateInventoryBatch(kind string, productID uuid.UUID, at time.Time) string {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	if kind == "" {
		kind = "STK"
	}
	productCode := strings.ToUpper(strings.ReplaceAll(productID.String()[:8], "-", ""))
	return fmt.Sprintf("%s-%s-%s-%s", kind, at.Format("20060102"), productCode, at.Format("150405"))
}
