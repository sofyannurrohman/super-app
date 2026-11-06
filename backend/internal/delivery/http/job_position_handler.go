package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type JobPositionHandler struct {
	useCase usecase.JobPositionUseCase
}

func NewJobPositionHandler(useCase usecase.JobPositionUseCase) *JobPositionHandler {
	return &JobPositionHandler{useCase: useCase}
}

func (h *JobPositionHandler) SetupRoutes(r *gin.RouterGroup) {
	positions := r.Group("/job-positions")
	{
		positions.POST("", h.Create)
		positions.GET("/:id", h.GetByID)
		positions.PUT("/:id", h.Update)
		positions.DELETE("/:id", h.Delete)
	}
}

func (h *JobPositionHandler) SetupPublicRoutes(r *gin.RouterGroup) {
	r.GET("/job-positions", h.GetAll)
}

func (h *JobPositionHandler) Create(c *gin.Context) {
	var req usecase.CreateJobPositionRequest
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

	if err := h.useCase.CreateJobPosition(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Job position created successfully"})
}

func (h *JobPositionHandler) GetAll(c *gin.Context) {
	positions, err := h.useCase.GetJobPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": positions})
}

func (h *JobPositionHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	position, err := h.useCase.GetJobPositionByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if position == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job position not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": position})
}

func (h *JobPositionHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req usecase.UpdateJobPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateJobPosition(id, &req); err != nil {
		if err.Error() == "job position not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job position updated successfully"})
}

func (h *JobPositionHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.useCase.DeleteJobPosition(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job position deleted successfully"})
}
