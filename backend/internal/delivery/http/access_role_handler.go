package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type AccessRoleHandler struct {
	usecase usecase.AccessRoleUsecase
}

func NewAccessRoleHandler(u usecase.AccessRoleUsecase) *AccessRoleHandler {
	return &AccessRoleHandler{u}
}

func (h *AccessRoleHandler) SetupRoutes(rg *gin.RouterGroup) {
	rolesGroup := rg.Group("/access-roles")
	{
		rolesGroup.GET("", h.GetAllRoles)
		rolesGroup.GET("/:id", h.GetRoleByID)
		rolesGroup.POST("", h.CreateRole)
		rolesGroup.PUT("/:id", h.UpdateRole)
		rolesGroup.DELETE("/:id", h.DeleteRole)
	}

	modulesGroup := rg.Group("/app-modules")
	{
		modulesGroup.GET("", h.GetAllModules)
		modulesGroup.POST("", h.CreateModule)
	}

	menusGroup := rg.Group("/app-menus")
	{
		menusGroup.GET("", h.GetAllMenus)
		menusGroup.POST("", h.CreateMenu)
	}
}

type createRoleRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	MenuIDs     []uuid.UUID `json:"menu_ids"`
}

func (h *AccessRoleHandler) CreateRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	role := &domain.AccessRole{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.usecase.CreateRole(c.Request.Context(), role, req.MenuIDs); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Role created successfully", role)
}

func (h *AccessRoleHandler) GetRoleByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Role ID")
		return
	}

	role, err := h.usecase.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Role retrieved successfully", role)
}

func (h *AccessRoleHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.usecase.GetAllRoles(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Roles retrieved successfully", roles)
}

func (h *AccessRoleHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Role ID")
		return
	}

	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	role, err := h.usecase.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Role not found")
		return
	}

	role.Name = req.Name
	role.Description = req.Description

	if err := h.usecase.UpdateRole(c.Request.Context(), role, req.MenuIDs); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Role updated successfully", role)
}

func (h *AccessRoleHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Role ID")
		return
	}

	if err := h.usecase.DeleteRole(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Role deleted successfully", nil)
}

type createModuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *AccessRoleHandler) CreateModule(c *gin.Context) {
	var req createModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	module := &domain.AppModule{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.usecase.CreateModule(c.Request.Context(), module); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Module created successfully", module)
}

func (h *AccessRoleHandler) GetAllModules(c *gin.Context) {
	modules, err := h.usecase.GetAllModules(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Modules retrieved successfully", modules)
}

type createMenuRequest struct {
	ModuleID uuid.UUID  `json:"module_id" binding:"required"`
	Name     string     `json:"name" binding:"required"`
	Path     string     `json:"path"`
	Level    int        `json:"level"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (h *AccessRoleHandler) CreateMenu(c *gin.Context) {
	var req createMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	menu := &domain.AppMenu{
		ModuleID: req.ModuleID,
		Name:     req.Name,
		Path:     req.Path,
		Level:    req.Level,
		ParentID: req.ParentID,
	}

	if err := h.usecase.CreateMenu(c.Request.Context(), menu); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Menu created successfully", menu)
}

func (h *AccessRoleHandler) GetAllMenus(c *gin.Context) {
	menus, err := h.usecase.GetAllMenus(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Menus retrieved successfully", menus)
}
