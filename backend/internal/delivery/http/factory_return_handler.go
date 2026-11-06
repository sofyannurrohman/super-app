package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type FactoryReturnHandler struct {
	returnUsecase usecase.FactoryReturnUsecase
}

func NewFactoryReturnHandler(u usecase.FactoryReturnUsecase) *FactoryReturnHandler {
	return &FactoryReturnHandler{u}
}

func (h *FactoryReturnHandler) RegisterRoutes(r *gin.RouterGroup) {
	returns := r.Group("/factory-returns")
	{
		returns.POST("", h.CreateReturn)
		returns.POST("/:id/receive", h.ReceiveReturn)
		returns.PUT("/:id/approve", h.ReceiveReturn)
		returns.POST("/:id/reject", h.RejectReturn)
		returns.PUT("/:id/reject", h.RejectReturn)
		returns.GET("/branch/:branchId", h.GetByBranch)
		returns.GET("/factory/:factoryId", h.GetByFactory)
		returns.GET("/company", h.GetByCompany)
	}
}

func (h *FactoryReturnHandler) CreateReturn(c *gin.Context) {
	var req usecase.CreateFactoryReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ret, err := h.returnUsecase.CreateReturn(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ret)
}

func (h *FactoryReturnHandler) ReceiveReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return id"})
		return
	}

	var req usecase.ReceiveFactoryReturnRequest
	// Optional body: if body is provided, bind it; otherwise ignore EOF
	_ = c.ShouldBindJSON(&req)

	if req.ReceiverName == "" {
		if val, ok := c.Get("employeeName"); ok && val != nil {
			req.ReceiverName = val.(string)
		} else if val, ok := c.Get("username"); ok && val != nil {
			req.ReceiverName = val.(string)
		}
	}
	if req.InspectedByName == "" {
		if val, ok := c.Get("employeeName"); ok && val != nil {
			req.InspectedByName = val.(string)
		} else if val, ok := c.Get("username"); ok && val != nil {
			req.InspectedByName = val.(string)
		}
	}

	if err := h.returnUsecase.ReceiveReturn(id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Return approved and received successfully"})
}

func (h *FactoryReturnHandler) RejectReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid return id"})
		return
	}

	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alasan penolakan wajib diisi"})
		return
	}

	var rejectorID *uuid.UUID
	if val, ok := c.Get("employeeID"); ok && val != nil {
		if eid, ok := val.(uuid.UUID); ok {
			rejectorID = &eid
		}
	}

	rejectorName := "Admin Nota / Pabrik"
	if val, ok := c.Get("employeeName"); ok && val != nil {
		rejectorName = val.(string)
	} else if val, ok := c.Get("username"); ok && val != nil {
		rejectorName = val.(string)
	}

	if err := h.returnUsecase.RejectReturn(id, rejectorID, rejectorName, body.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Retur berhasil ditolak"})
}

func (h *FactoryReturnHandler) GetByBranch(c *gin.Context) {
	branchIdStr := c.Param("branchId")
	if branchIdStr == "ALL_BRANCHES" {
		val, exists := c.Get("companyID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no companyID in context"})
			return
		}

		var companyID uuid.UUID
		var err error

		if id, ok := val.(uuid.UUID); ok {
			companyID = id
		} else if idStr, ok := val.(string); ok {
			companyID, err = uuid.Parse(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID format"})
				return
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Company ID type not recognized"})
			return
		}

		returns, err := h.returnUsecase.GetByCompany(companyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, returns)
		return
	}

	id, err := uuid.Parse(branchIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}

	returns, err := h.returnUsecase.GetByBranch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, returns)
}

func (h *FactoryReturnHandler) GetByFactory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("factoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid factory id"})
		return
	}

	returns, err := h.returnUsecase.GetByFactory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, returns)
}

func (h *FactoryReturnHandler) GetByCompany(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no companyID in context"})
		return
	}

	var companyID uuid.UUID
	var err error

	if id, ok := val.(uuid.UUID); ok {
		companyID = id
	} else if idStr, ok := val.(string); ok {
		companyID, err = uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID format"})
			return
		}
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company ID type not recognized"})
		return
	}

	returns, err := h.returnUsecase.GetByCompany(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, returns)
}
