package http

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"github.com/xuri/excelize/v2"
)

type ImportPreviewResult struct {
	OriginalRow     int           `json:"original_row"`
	Name            string        `json:"name"`
	Address         string        `json:"address"`
	PhoneNumber     string        `json:"phone_number"`
	DefaultSalesman string        `json:"default_salesman"`
	Status          string        `json:"status"` // "New", "Duplicate", "Potential Duplicate"
	DuplicateOf     *domain.Store `json:"duplicate_of"`
}

type StoreImportRequest struct {
	Name            string `json:"name"`
	Address         string `json:"address"`
	PhoneNumber     string `json:"phone_number"`
	DefaultSalesman string `json:"default_salesman"`
	Action          string `json:"action"`          // "Create", "Update", "Skip"
	DuplicateOfID   string `json:"duplicate_of_id"` // ID if Update
}

type ImportProcessRequest struct {
	CompanyID string               `json:"company_id"`
	Stores    []StoreImportRequest `json:"stores"`
}

// POST /api/v1/stores/import/preview
func (h *StoreHandler) ImportPreview(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	companyIDStr := c.PostForm("company_id")
	var companyID *uuid.UUID
	if companyIDStr != "" {
		parsed, err := uuid.Parse(companyIDStr)
		if err == nil {
			companyID = &parsed
		}
	} else {
		// Try from user's employee profile
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil && employee.CompanyID != nil {
				companyID = employee.CompanyID
			}
		}
	}

	// Fetch existing stores for comparison (lightweight without heavy financial subqueries)
	existingStores, _, err := h.storeRepo.FindPaginated(companyID, nil, nil, "", "", "ALL", "", 100000, 0, "", "", false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing stores"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Excel file"})
		return
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file is empty"})
		return
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read rows"})
		return
	}

	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No data rows found"})
		return
	}

	// Dynamic header mapping
	headers := rows[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	idxName := getColIndex(headerMap, "nama", "pelanggan")
	idxAddress := getColIndex(headerMap, "alamat")
	idxPhone := getColIndex(headerMap, "telepon", "telp")
	idxSales := getColIndex(headerMap, "penjual", "sales")

	if idxName == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kolom Nama Pelanggan tidak ditemukan"})
		return
	}

	var results []ImportPreviewResult

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		name := getCellValue(row, idxName)
		if strings.TrimSpace(name) == "" {
			continue // Skip empty name
		}

		addr := getCellValue(row, idxAddress)
		phone := getCellValue(row, idxPhone)
		sales := getCellValue(row, idxSales)

		status, duplicateOf := determineImportStatus(name, phone, existingStores)

		results = append(results, ImportPreviewResult{
			OriginalRow:     i + 1,
			Name:            name,
			Address:         addr,
			PhoneNumber:     phone,
			DefaultSalesman: sales,
			Status:          status,
			DuplicateOf:     duplicateOf,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Preview generated",
		"data":    results,
	})
}

// POST /api/v1/stores/import
func (h *StoreHandler) ImportProcess(c *gin.Context) {
	var req ImportProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var compID uuid.UUID
	if req.CompanyID != "" {
		compID, _ = uuid.Parse(req.CompanyID)
	} else {
		// Use from user context
		if userIDStr, exists := c.Get("userID"); exists {
			userID := userIDStr.(uuid.UUID)
			employee, err := h.employeeUseCase.GetEmployeeByUserID(userID)
			if err == nil && employee != nil && employee.CompanyID != nil {
				compID = *employee.CompanyID
			}
		}
	}

	createdCount := 0
	updatedCount := 0

	for _, item := range req.Stores {
		if item.Action == "Skip" {
			continue
		}

		if item.Action == "Create" {
			newStore := domain.Store{
				CompanyID:   compID,
				Name:        item.Name,
				Address:     item.Address,
				PhoneNumber: item.PhoneNumber,
				IsActive:    true,
				Category:    "TOKO_BARU",
			}
			
			// Optional: Try to find Employee ID from item.DefaultSalesman here if needed
			// For now, left nil (auto assign logic may apply in Repo or manual mapping)

			if err := h.storeRepo.Create(&newStore); err == nil {
				createdCount++
				if newStore.Address != "" {
					go h.backgroundGeocodeStoreByAddress(newStore.ID, newStore.Address)
				}
			}
		} else if item.Action == "Update" && item.DuplicateOfID != "" {
			existingID, err := uuid.Parse(item.DuplicateOfID)
			if err == nil {
				existing, err := h.storeRepo.FindByID(existingID)
				if err == nil && existing != nil {
					// We only update if empty or if requested
					if existing.PhoneNumber == "" && item.PhoneNumber != "" {
						existing.PhoneNumber = item.PhoneNumber
					}
					if existing.Address == "" && item.Address != "" {
						existing.Address = item.Address
					}
					// Not overwriting Name because it's usually already correct in DB
					if err := h.storeRepo.Update(existing); err == nil {
						updatedCount++
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Import successful. Created %d stores, Updated %d stores.", createdCount, updatedCount),
		"data": gin.H{
			"created": createdCount,
			"updated": updatedCount,
		},
	})
}

// -- Helpers --

func getCellValue(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func getColIndex(headerMap map[string]int, keywords ...string) int {
	for k, idx := range headerMap {
		for _, kw := range keywords {
			if strings.Contains(k, kw) {
				return idx
			}
		}
	}
	return -1
}

func determineImportStatus(name, phone string, existingStores []domain.Store) (string, *domain.Store) {
	normPhone := normalizePhone(phone)
	normName := normalizeName(name)

	var bestMatch *domain.Store
	bestScore := 0.0

	for i := range existingStores {
		dbStore := &existingStores[i]

		// 1. Exact phone match
		if normPhone != "" && normalizePhone(dbStore.PhoneNumber) == normPhone {
			return "Duplicate", dbStore
		}

		// 2. Exact name match (normalized)
		if normName != "" && normalizeName(dbStore.Name) == normName {
			return "Duplicate", dbStore
		}

		// 3. Fuzzy name match
		if normName != "" {
			score := similarity(normName, normalizeName(dbStore.Name))
			if score > bestScore {
				bestScore = score
				bestMatch = dbStore
			}
		}
	}

	if bestScore >= 0.8 {
		return "Potential Duplicate", bestMatch
	}

	return "New", nil
}

func normalizePhone(p string) string {
	re := regexp.MustCompile(`[^0-9]`)
	p = re.ReplaceAllString(p, "")
	
	// Example standardization (08 -> 628)
	if strings.HasPrefix(p, "08") {
		p = "628" + p[2:]
	}
	return p
}

func normalizeName(n string) string {
	n = strings.ToLower(strings.TrimSpace(n))
	prefixes := []string{"tk ", "toko ", "wr ", "warung ", "cv ", "pt "}
	for _, p := range prefixes {
		if strings.HasPrefix(n, p) {
			n = strings.TrimPrefix(n, p)
		}
	}
	re := regexp.MustCompile(`[^a-z0-9 ]`)
	n = re.ReplaceAllString(n, "")
	return strings.TrimSpace(n)
}

func similarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	dist := levenshtein(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}

func levenshtein(a, b string) int {
	aRune := []rune(a)
	bRune := []rune(b)

	if len(aRune) == 0 {
		return len(bRune)
	}
	if len(bRune) == 0 {
		return len(aRune)
	}

	row := make([]int, len(bRune)+1)
	for i := 0; i <= len(bRune); i++ {
		row[i] = i
	}

	for i := 1; i <= len(aRune); i++ {
		prev := i
		for j := 1; j <= len(bRune); j++ {
			cost := 0
			if aRune[i-1] != bRune[j-1] {
				cost = 1
			}
			minCost := prev + 1
			if row[j]+1 < minCost {
				minCost = row[j] + 1
			}
			if row[j-1]+cost < minCost {
				minCost = row[j-1] + cost
			}
			row[j-1] = prev
			prev = minCost
		}
		row[len(bRune)] = prev
	}
	return row[len(bRune)]
}
