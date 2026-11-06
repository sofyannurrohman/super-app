package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type SalesReturnHandler struct {
	returnUsecase   usecase.SalesReturnUsecase
	employeeUseCase usecase.EmployeeUsecase
}

func NewSalesReturnHandler(ru usecase.SalesReturnUsecase, eu usecase.EmployeeUsecase) *SalesReturnHandler {
	return &SalesReturnHandler{ru, eu}
}

func (h *SalesReturnHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/admin/sales-returns", h.GetByBranch)
	router.GET("/admin/sales-returns/:id", h.GetByID)
	router.POST("/admin/sales-returns", h.CreateReturn)
	router.PUT("/admin/sales-returns/:id", h.UpdateReturn)
	router.DELETE("/admin/sales-returns/:id", h.DeleteReturn)
	router.PATCH("/admin/sales-returns/:id/approve", h.ApproveReturnNota) // keep for backward compatibility or Admin Nota
	router.PATCH("/admin/sales-returns/:id/approve-nota", h.ApproveReturnNota)
	router.PATCH("/admin/sales-returns/:id/approve-warehouse", h.ApproveReturnWarehouse)
}

func (h *SalesReturnHandler) CreateReturn(c *gin.Context) {
	var req usecase.CreateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve EmployeeID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			req.EmployeeID = emp.ID
		}
	}

	ret, err := h.returnUsecase.CreateReturn(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": ret})
}

func (h *SalesReturnHandler) ApproveReturnNota(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var approvedByID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			approvedByID = emp.ID
		}
	}

	if err := h.returnUsecase.ApproveReturnByNota(id, approvedByID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Retur berhasil diproses, tagihan disesuaikan. Menunggu persetujuan fisik Gudang."})
}

func (h *SalesReturnHandler) ApproveReturnWarehouse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var approvedByID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID := userIDStr.(uuid.UUID)
		emp, err := h.employeeUseCase.GetEmployeeByUserID(userID)
		if err == nil && emp != nil {
			approvedByID = emp.ID
		}
	}

	if err := h.returnUsecase.ApproveReturnByWarehouse(id, approvedByID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Retur fisik berhasil diterima, stok gudang bertambah."})
}

func (h *SalesReturnHandler) GetByBranch(c *gin.Context) {
	branchID, _ := uuid.Parse(c.Query("branch_id"))
	status := c.Query("status")
	returns, err := h.returnUsecase.GetByBranch(branchID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": returns})
}

func (h *SalesReturnHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	ret, err := h.returnUsecase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Retur tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ret})
}

func (h *SalesReturnHandler) UpdateReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.UpdateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ret, err := h.returnUsecase.UpdateReturn(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ret})
}

func (h *SalesReturnHandler) DeleteReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.returnUsecase.DeleteReturn(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Retur berhasil dihapus"})
}
