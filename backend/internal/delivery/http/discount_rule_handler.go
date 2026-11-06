package http

import (
	"github.com/gin-gonic/gin"

	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"net/http"
)

type DiscountRuleHandler struct {
	db *gorm.DB
}

func NewDiscountRuleHandler(r *gin.Engine, db *gorm.DB) {
	handler := &DiscountRuleHandler{db: db}

	api := r.Group("/api/v1/discount-rules")
	{
		api.GET("", handler.GetAll)
		api.POST("", handler.Create)
		api.PUT("/:id", handler.Update)
		api.DELETE("/:id", handler.Delete)
	}
}

func (h *DiscountRuleHandler) GetAll(c *gin.Context) {
	var rules []domain.DiscountRule
	if err := h.db.Preload("Tiers").Preload("ExcludedProducts.Product").Preload("ValidUnits").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *DiscountRuleHandler) Create(c *gin.Context) {
	var req domain.DiscountRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *DiscountRuleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req domain.DiscountRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing domain.DiscountRule
	if err := h.db.First(&existing, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Discount Rule not found"})
		return
	}

	// Update existing record
	// For simplicity, we just save the basic fields here. In production,
	// you should delete and recreate the nested relations (Tiers, ValidUnits) if they change.
	existing.Name = req.Name
	existing.IsActive = req.IsActive
	existing.PaymentMethodRequired = req.PaymentMethodRequired
	existing.StartDate = req.StartDate
	existing.EndDate = req.EndDate

	if err := h.db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}

func (h *DiscountRuleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// Delete relations first manually or set gorm constraint to CASCADE
	h.db.Where("discount_rule_id = ?", id).Delete(&domain.DiscountTier{})
	h.db.Where("discount_rule_id = ?", id).Delete(&domain.DiscountExcludedProduct{})
	h.db.Where("discount_rule_id = ?", id).Delete(&domain.DiscountValidUnit{})

	if err := h.db.Delete(&domain.DiscountRule{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Discount Rule deleted"})
}
