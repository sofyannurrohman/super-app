package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type EpkwtHandler struct {
	useCase usecase.EpkwtUseCase
}

func NewEpkwtHandler(u usecase.EpkwtUseCase) *EpkwtHandler {
	return &EpkwtHandler{useCase: u}
}

func (h *EpkwtHandler) RegisterRoutes(r *gin.RouterGroup) {
	epkwt := r.Group("/epkwt")
	{
		epkwt.GET("", h.GetAll)
		epkwt.GET("/:id", h.GetByID)
		epkwt.POST("", h.Create)
		epkwt.PUT("/:id", h.Update)
		epkwt.PUT("/:id/sign", h.Sign)
		epkwt.POST("/:id/hr-signature", h.UploadHrSignature)
		epkwt.POST("/:id/document", h.UploadDocument)
	}
}

func (h *EpkwtHandler) Create(c *gin.Context) {
	roleVal, _ := c.Get("role")
	role := ""
	if roleVal != nil {
		role = strings.ToLower(roleVal.(string))
	}

	var req struct {
		usecase.CreateEpkwtRequestParams
		EmployeeID *uuid.UUID `json:"employee_id"`
		SignatureData string `json:"signature_data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var targetEmployeeID uuid.UUID
	if role == "superadmin" || role == "hr_admin" || role == "admin" {
		if req.EmployeeID != nil {
			targetEmployeeID = *req.EmployeeID
		} else {
			empIDVal, exists := c.Get("employeeID")
			if !exists {
				utils.ErrorResponse(c, http.StatusBadRequest, "employee_id is required for admin role")
				return
			}
			targetEmployeeID = empIDVal.(uuid.UUID)
		}
	} else {
		empIDVal, exists := c.Get("employeeID")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "Hanya karyawan terdaftar yang bisa mengajukan EPKWT")
			return
		}
		targetEmployeeID = empIDVal.(uuid.UUID)
	}

	// Process SignatureData if provided
	if req.SignatureData != "" {
		b64 := req.SignatureData
		if idx := strings.Index(b64, ","); idx != -1 {
			b64 = b64[idx+1:]
		}

		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err == nil {
			uploadDir := "./uploads/signatures"
			os.MkdirAll(uploadDir, 0755)
			filename := fmt.Sprintf("sig_req_%s_%d.png", targetEmployeeID.String(), time.Now().Unix())
			filePath := filepath.Join(uploadDir, filename)
			if err := os.WriteFile(filePath, decoded, 0644); err == nil {
				url := "/uploads/signatures/" + filename
				req.EmployeeSignatureURL = &url
			}
		}
	}

	epkwt, err := h.useCase.CreateRequest(targetEmployeeID, &req.CreateEpkwtRequestParams)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Pengajuan EPKWT berhasil dibuat", epkwt)
}

func (h *EpkwtHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var req usecase.UpdateEpkwtRequestParams
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.useCase.UpdateRequest(id, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "EPKWT berhasil diperbarui", nil)
}

type SignRequestParams struct {
	SignatureData string `json:"signature_data"` // base64 string
	SignatureURL  string `json:"signature_url"`  // optional URL
}

func (h *EpkwtHandler) Sign(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var req SignRequestParams
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	signatureURL := req.SignatureURL

	// Decode base64 image if signature_data is supplied
	if req.SignatureData != "" {
		// Clean up prefix if any (e.g. data:image/png;base64,)
		b64 := req.SignatureData
		if idx := strings.Index(b64, ","); idx != -1 {
			b64 = b64[idx+1:]
		}

		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Format signature_data base64 tidak valid")
			return
		}

		// Ensure directory exists
		uploadDir := "./uploads/signatures"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat folder tanda tangan")
			return
		}

		filename := fmt.Sprintf("sig_%s_%d.png", id.String(), time.Now().Unix())
		filePath := filepath.Join(uploadDir, filename)

		if err := os.WriteFile(filePath, decoded, 0644); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan file tanda tangan")
			return
		}

		signatureURL = "/uploads/signatures/" + filename
	}

	if signatureURL == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tanda tangan diperlukan (signature_data atau signature_url)")
		return
	}

	if err := h.useCase.SignRequest(id, signatureURL); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "EPKWT berhasil ditandatangani", gin.H{"signature_url": signatureURL})
}

func (h *EpkwtHandler) UploadHrSignature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File gambar diperlukan")
		return
	}

	uploadDir := "./uploads/signatures"
	os.MkdirAll(uploadDir, 0755)

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("hr_sig_%s_%d%s", id.String(), time.Now().Unix(), ext)
	savePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan file")
		return
	}

	url := "/uploads/signatures/" + filename
	req := usecase.UpdateEpkwtRequestParams{
		HrSignatureURL: &url,
	}

	if err := h.useCase.UpdateRequest(id, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tanda tangan HRD berhasil diunggah", gin.H{"url": url})
}

func (h *EpkwtHandler) UploadDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File dokumen diperlukan")
		return
	}

	// Max 15MB
	if file.Size > 15*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Ukuran file maksimal 15MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format file harus berupa PDF, JPG, atau PNG")
		return
	}

	uploadDir := "./uploads/epkwt_documents"
	os.MkdirAll(uploadDir, 0755)

	filename := fmt.Sprintf("epkwt_doc_%s_%d%s", id.String(), time.Now().Unix(), ext)
	savePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan file")
		return
	}

	url := "/uploads/epkwt_documents/" + filename
	if err := h.useCase.UploadDocument(id, url); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Dokumen EPKWT berhasil diunggah", gin.H{"url": url})
}

func (h *EpkwtHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	req, err := h.useCase.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data EPKWT tidak ditemukan")
		return
	}

	roleVal, _ := c.Get("role")
	if roleVal != nil && strings.ToLower(roleVal.(string)) == "employee" {
		empIDVal, exists := c.Get("employeeID")
		if !exists || req.EmployeeID != empIDVal.(uuid.UUID) {
			utils.ErrorResponse(c, http.StatusForbidden, "Akses ditolak: Anda hanya bisa melihat data Anda sendiri")
			return
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Data EPKWT berhasil diambil", req)
}

func (h *EpkwtHandler) GetAll(c *gin.Context) {
	roleVal, _ := c.Get("role")
	role := ""
	if roleVal != nil {
		role = strings.ToLower(roleVal.(string))
	}

	var employeeID *uuid.UUID

	// If employee, only show their own PKWTs
	if role == "employee" {
		empIDVal, exists := c.Get("employeeID")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, "Akses ditolak")
			return
		}
		eID := empIDVal.(uuid.UUID)
		employeeID = &eID
	} else {
		// Admin / HR can filter by employee_id query param
		if empQuery := c.Query("employee_id"); empQuery != "" {
			if parsed, err := uuid.Parse(empQuery); err == nil {
				employeeID = &parsed
			}
		}
	}

	reqs, err := h.useCase.GetAll(employeeID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar EPKWT berhasil diambil", reqs)
}
