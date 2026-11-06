package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type ShiftHandler struct {
	useCase usecase.ShiftUseCase
}

func NewShiftHandler(useCase usecase.ShiftUseCase) *ShiftHandler {
	return &ShiftHandler{useCase: useCase}
}

func (h *ShiftHandler) SetupRoutes(r *gin.RouterGroup) {
	shifts := r.Group("/shifts")
	{
		shifts.POST("", h.Create)
		shifts.GET("/:id", h.GetByID)
		shifts.PUT("/:id", h.Update)
		shifts.DELETE("/:id", h.Delete)
	}
}

func (h *ShiftHandler) SetupPublicRoutes(r *gin.RouterGroup) {
	r.GET("/shifts", h.GetAll)
}

func (h *ShiftHandler) Create(c *gin.Context) {
	var req usecase.CreateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Automatically use company ID from JWT claims
	if companyID, exists := c.Get("companyID"); exists {
		cid := companyID.(uuid.UUID)
		if cid != uuid.Nil {
			req.CompanyID = &cid
		}
	}

	if err := h.useCase.CreateShift(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Shift created successfully"})
}

func (h *ShiftHandler) GetAll(c *gin.Context) {
	shifts, err := h.useCase.GetShifts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shifts})
}

func (h *ShiftHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	shift, err := h.useCase.GetShiftByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if shift == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shift not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shift})
}

func (h *ShiftHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req usecase.UpdateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateShift(id, &req); err != nil {
		if err.Error() == "shift not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Shift updated successfully"})
}

func (h *ShiftHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.useCase.DeleteShift(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Shift deleted successfully"})
}
