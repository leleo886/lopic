package admin_controllers

import (
	"net/http"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/services/admin_services"
)

type RoleController struct {
	roleService *admin_services.RoleService
}

func NewRoleController(roleService *admin_services.RoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

// GetRoles 获取角色列表
// @Summary 获取角色列表
// @Description 获取所有角色，支持分页
// @Tags 角色管理员
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param searchkey query string false "搜索关键词"
// @Param orderby query string false "排序字段" default(created_at)
// @Param order query string false "排序方向" default(desc)
// @Success 200 {object} success.DataResponse{data=[]models.Role}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/roles [get]
func (h *RoleController) GetRoles(c *gin.Context) {
	page, err1 := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err2 := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	searchkey := c.DefaultQuery("searchkey", "")
	orderby := c.DefaultQuery("orderby", "created_at")
	order := c.DefaultQuery("order", "desc")

	if err1 != nil || err2 != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	if order != "asc" {
		order = "desc"
	}
	orderby = strings.TrimSpace(orderby)
	order = strings.TrimSpace(order)

	offset := (page - 1) * pageSize

	roles, err := h.roleService.GetRoles(page, pageSize, offset, searchkey, orderby, order)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Roles retrieved successfully", roles))
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 200 {object} success.DataResponse{data=models.Role}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/roles/{id} [get]
func (h *RoleController) GetRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	role, err := h.roleService.GetRole(id)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Role retrieved successfully", role))
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建一个新角色
// @Tags 角色管理员
// @Produce json
// @Security ApiKeyAuth
// @Param role body admin_services.RoleRequest true "角色信息"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/roles [post]
func (h *RoleController) CreateRole(c *gin.Context) {
	var role admin_services.RoleRequest
	if err := c.ShouldBindJSON(&role); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if err := roleRequestCheck(&role); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	err := h.roleService.CreateRole(&models.Role{
		Name:              role.Name,
		Description:       role.Description,
		AllowedExtensions: role.AllowedExtensions,
		MaxFilesPerUpload: role.MaxFilesPerUpload,
		MaxFileSizeMB:     role.MaxFileSizeMB,
		MaxAlbumsPerUser:  role.MaxAlbumsPerUser,
		MaxStorageSizeMB:  role.MaxStorageSizeMB,
		GalleryOpen:       role.GalleryOpen,
		StorageName:       role.StorageName,
	})
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Role created successfully"))
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 根据ID更新角色信息
// @Tags 角色管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Param role body admin_services.RoleRequest true "角色信息"
// @Success 200 {object} success.DataResponse{data=models.Role}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/roles/{id} [put]
func (h *RoleController) UpdateRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	var role admin_services.RoleRequest
	if err := c.ShouldBindJSON(&role); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if err := roleRequestCheck(&role); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	err = h.roleService.UpdateRole(uint(id), &role)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Role updated successfully", role))

}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 根据ID删除角色
// @Tags 角色管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "角色ID"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/roles/{id} [delete]
func (h *RoleController) DeleteRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	err = h.roleService.DeleteRole(uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewSuccessResponse("Role deleted successfully"))
}

// GetUsersCountByRole 获取不同角色的用户数量
// @Summary 获取不同角色的用户数量
// @Description 获取每个角色下的用户数量
// @Tags 角色管理员
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} success.DataResponse{data=map[string]int}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/roles/users-count [get]
func (h *RoleController) GetUsersCountByRole(c *gin.Context) {
	counts, err := h.roleService.GetUsersCountByRole()
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Users count by role retrieved successfully", counts))
}

func roleRequestCheck(role *admin_services.RoleRequest) error {
	if role.Name == "" {
		return cerrors.ErrBadRequest
	}
	if role.MaxFilesPerUpload < -1 || role.MaxFileSizeMB < -1 || role.MaxAlbumsPerUser < -1 || role.MaxStorageSizeMB < -1 {
		return cerrors.ErrBadRequest
	}
	return nil
}
