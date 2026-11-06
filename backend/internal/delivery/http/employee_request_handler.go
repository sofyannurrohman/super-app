package http

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/internal/usecase"
)

type EmployeeRequestHandler struct {
	useCase  usecase.EmployeeRequestUseCase
	empRepo  repository.EmployeeRepository
}

func NewEmployeeRequestHandler(useCase usecase.EmployeeRequestUseCase, empRepo repository.EmployeeRepository) *EmployeeRequestHandler {
	return &EmployeeRequestHandler{
		useCase: useCase,
		empRepo: empRepo,
	}
}

func (h *EmployeeRequestHandler) SetupRoutes(r *gin.RouterGroup) {
	group := r.Group("/employee-requests")
	{
		group.GET("", h.GetAll)
		group.GET("/summary", h.GetSummary)
		group.GET("/:id", h.GetByID)
		group.POST("", h.Create)
		group.POST("/upload-cv", h.UploadCandidateCV)
		group.POST("/:id/candidates", h.AddCandidates)
		group.DELETE("/:id/candidates/:candidate_id", h.DeleteCandidate)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
		group.PUT("/:id/approve", h.Approve)
		group.PUT("/:id/approve-hr", h.ApproveHR)
		group.PUT("/:id/approve-director", h.ApproveDirector)
		group.PUT("/:id/reject", h.Reject)
		group.PUT("/:id/progress", h.UpdateProgress)
	}
}

func (h *EmployeeRequestHandler) Create(c *gin.Context) {
	var payload usecase.CreateEmployeeRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approverInfo := GetApproverInfo(c, h.empRepo)

	var requesterID *uuid.UUID = approverInfo.EmployeeID
	requesterName := approverInfo.FullName
	requesterJobTitle := approverInfo.JobPosition

	if requesterName == "" || requesterName == "Admin / Approver" {
		if val, exists := c.Get("name"); exists {
			if n, ok := val.(string); ok && n != "" {
				requesterName = n
			}
		}
	}
	if requesterJobTitle == "" {
		requesterJobTitle = "Kepala Divisi"
	}

	res, err := h.useCase.Create(&payload, requesterID, requesterName, requesterJobTitle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Permintaan karyawan berhasil diajukan",
		"data":    res,
	})
}

func (h *EmployeeRequestHandler) UploadCandidateCV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File dokumen / CV wajib diunggah"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]string{
		".pdf":  "PDF",
		".jpg":  "IMAGE",
		".jpeg": "IMAGE",
		".png":  "IMAGE",
	}

	docType, allowed := allowedExtensions[ext]
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung (gunakan PDF, JPG, JPEG, atau PNG)"})
		return
	}

	// Limit PDF size to 10MB
	if ext == ".pdf" && file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file PDF maksimal 10 MB"})
		return
	}

	uploadDir := "./uploads/candidate_cvs"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat direktori upload"})
		return
	}

	filename := fmt.Sprintf("cv_%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	if docType == "IMAGE" {
		// Standardize compressed images to .jpg
		filename = fmt.Sprintf("cv_%s_%d.jpg", uuid.New().String()[:8], time.Now().Unix())
	}
	savePath := filepath.Join(uploadDir, filename)

	if docType == "IMAGE" {
		// Compress image
		if err := compressAndSaveImage(file, savePath); err != nil {
			// Fallback to direct save if compression fails
			if saveErr := c.SaveUploadedFile(file, savePath); saveErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file gambar"})
				return
			}
		}
	} else {
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file CV"})
			return
		}
	}

	fileURL := "/uploads/candidate_cvs/" + filename

	c.JSON(http.StatusOK, gin.H{
		"message":       "File berhasil diunggah",
		"document_url":  fileURL,
		"document_name": file.Filename,
		"document_type": docType,
	})
}

func compressAndSaveImage(fileHeader *multipart.FileHeader, destPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}

	// Resize if dimension exceeds 1600px
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	maxWidth, maxHeight := 1600, 1600

	if width > maxWidth || height > maxHeight {
		var newW, newH int
		if width > height {
			newW = maxWidth
			newH = (height * maxWidth) / width
		} else {
			newH = maxHeight
			newW = (width * maxHeight) / height
		}

		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		// Use bilinear interpolation from golang.org/x/image/draw or standard sub-sampling
		drawNearestNeighbor(dst, img, newW, newH)
		img = dst
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Encode to JPEG with 75% quality for optimal balance of size and crispness
	return jpeg.Encode(outFile, img, &jpeg.Options{Quality: 75})
}

func drawNearestNeighbor(dst *image.RGBA, src image.Image, targetWidth, targetHeight int) {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	for y := 0; y < targetHeight; y++ {
		srcY := srcBounds.Min.Y + (y * srcH / targetHeight)
		for x := 0; x < targetWidth; x++ {
			srcX := srcBounds.Min.X + (x * srcW / targetWidth)
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
}

func (h *EmployeeRequestHandler) AddCandidates(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var payload usecase.AddCandidatesPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.useCase.AddCandidates(id, &payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data karyawan baru berhasil ditambahkan",
		"data":    res,
	})
}

func (h *EmployeeRequestHandler) DeleteCandidate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID pengajuan tidak valid"})
		return
	}

	candidateIdParam := c.Param("candidate_id")
	candidateID, err := uuid.Parse(candidateIdParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID kandidat tidak valid"})
		return
	}

	res, err := h.useCase.DeleteCandidate(id, candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data karyawan berhasil dihapus dari pengajuan",
		"data":    res,
	})
}

func (h *EmployeeRequestHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var payload usecase.UpdateEmployeeRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.useCase.Update(id, &payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permintaan karyawan berhasil diperbarui",
		"data":    res,
	})
}

func (h *EmployeeRequestHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	if err := h.useCase.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permintaan karyawan berhasil dihapus"})
}

func (h *EmployeeRequestHandler) GetAll(c *gin.Context) {
	filter := repository.EmployeeRequestFilter{
		Status: c.Query("status"),
		Search: c.Query("search"),
	}

	if branchIDStr := c.Query("branch_id"); branchIDStr != "" && branchIDStr != "ALL" && branchIDStr != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(branchIDStr); err == nil {
			filter.BranchID = &parsed
		}
	}

	if deptIDStr := c.Query("department_id"); deptIDStr != "" && deptIDStr != "ALL" {
		if parsed, err := uuid.Parse(deptIDStr); err == nil {
			filter.DepartmentID = &parsed
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &t
		}
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	data, total, err := h.useCase.GetAll(filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *EmployeeRequestHandler) GetSummary(c *gin.Context) {
	var branchID *uuid.UUID
	var deptID *uuid.UUID

	if branchIDStr := c.Query("branch_id"); branchIDStr != "" && branchIDStr != "ALL" && branchIDStr != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &parsed
		}
	}
	if deptIDStr := c.Query("department_id"); deptIDStr != "" && deptIDStr != "ALL" {
		if parsed, err := uuid.Parse(deptIDStr); err == nil {
			deptID = &parsed
		}
	}

	stats, err := h.useCase.GetSummaryStats(branchID, deptID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *EmployeeRequestHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	res, err := h.useCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permintaan karyawan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

// ApproveHR handles Stage 1 Approval by HR Manager
func (h *EmployeeRequestHandler) ApproveHR(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var reqBody struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&reqBody)

	approverInfo := GetApproverInfo(c, h.empRepo)
	var approverID uuid.UUID
	if approverInfo.EmployeeID != nil {
		approverID = *approverInfo.EmployeeID
	} else {
		approverID = approverInfo.UserID
	}

	res, err := h.useCase.ApproveHR(id, approverID, reqBody.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Persetujuan HR Manager berhasil disimpan",
		"data":    res,
	})
}

// ApproveDirector handles Stage 2 Approval by Director
func (h *EmployeeRequestHandler) ApproveDirector(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var reqBody struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&reqBody)

	approverInfo := GetApproverInfo(c, h.empRepo)
	var approverID uuid.UUID
	if approverInfo.EmployeeID != nil {
		approverID = *approverInfo.EmployeeID
	} else {
		approverID = approverInfo.UserID
	}

	res, err := h.useCase.ApproveDirector(id, approverID, reqBody.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Persetujuan Direktur berhasil disimpan",
		"data":    res,
	})
}

// Approve handles generic/full approval (e.g. Superadmin)
func (h *EmployeeRequestHandler) Approve(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var reqBody struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&reqBody)

	approverInfo := GetApproverInfo(c, h.empRepo)
	var approverID uuid.UUID
	if approverInfo.EmployeeID != nil {
		approverID = *approverInfo.EmployeeID
	} else {
		approverID = approverInfo.UserID
	}

	res, err := h.useCase.Approve(id, approverID, reqBody.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permintaan karyawan berhasil disetujui penuh",
		"data":    res,
	})
}

func (h *EmployeeRequestHandler) Reject(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var reqBody struct {
		Reason string `json:"reason" binding:"required"`
		Role   string `json:"role"` // HR_MANAGER or DIRECTOR
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alasan penolakan wajib diisi"})
		return
	}

	approverInfo := GetApproverInfo(c, h.empRepo)
	var approverID uuid.UUID
	if approverInfo.EmployeeID != nil {
		approverID = *approverInfo.EmployeeID
	} else {
		approverID = approverInfo.UserID
	}

	if reqBody.Role == "" {
		reqBody.Role = "HR_MANAGER"
	}

	res, err := h.useCase.Reject(id, approverID, reqBody.Reason, reqBody.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permintaan karyawan berhasil ditolak",
		"data":    res,
	})
}

func (h *EmployeeRequestHandler) UpdateProgress(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var payload usecase.UpdateProgressPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.useCase.UpdateProgress(id, &payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Progres rekrutmen berhasil diperbarui",
		"data":    res,
	})
}
