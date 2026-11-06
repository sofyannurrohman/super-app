package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type BranchHandler struct {
	useCase usecase.BranchUseCase
}

func NewBranchHandler(useCase usecase.BranchUseCase) *BranchHandler {
	return &BranchHandler{useCase: useCase}
}

func (h *BranchHandler) SetupRoutes(r *gin.RouterGroup) {
	branches := r.Group("/branches")
	{
		branches.POST("", h.Create)
		branches.GET("/:id", h.GetByID)
		branches.PUT("/:id", h.Update)
		branches.DELETE("/:id", h.Delete)
		branches.GET("/:id/products/prices", h.GetProductPrices)
		branches.PUT("/:id/products/:product_id/prices", h.SetProductPrice)
	}
}

func (h *BranchHandler) SetupPublicRoutes(r *gin.RouterGroup) {
	r.GET("/branches", h.GetAll)
}

func (h *BranchHandler) Create(c *gin.Context) {
	var req usecase.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.CreateBranch(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Branch created successfully"})
}

func (h *BranchHandler) GetAll(c *gin.Context) {
	// Try to filter by company_id from JWT context first, then from query param
	var companyID uuid.UUID
	// Skip company_id filter for universal view in management
	/*
		if val, ok := c.Get("company_id"); ok && val != "" {
			if id, err := uuid.Parse(val.(string)); err == nil {
				companyID = id
			}
		} else if cid := c.Query("company_id"); cid != "" {
			if id, err := uuid.Parse(cid); err == nil {
				companyID = id
			}
		}
	*/

	var branches interface{}
	var err error
	if companyID != uuid.Nil {
		branches, err = h.useCase.GetBranchesByCompany(companyID)
	} else {
		branches, err = h.useCase.GetBranches()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": branches})
}

func (h *BranchHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	branch, err := h.useCase.GetBranchByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if branch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": branch})
}

func (h *BranchHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req usecase.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateBranch(id, &req); err != nil {
		if err.Error() == "branch not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch updated successfully"})
}

func (h *BranchHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.useCase.DeleteBranch(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch deleted successfully"})
}

func (h *BranchHandler) GetProductPrices(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	prices, err := h.useCase.GetProductPrices(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prices})
}

func (h *BranchHandler) SetProductPrice(c *gin.Context) {
	branchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req struct {
		HET float64 `json:"het" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.SetProductPrice(branchID, productID, req.HET); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product price updated successfully"})
}
