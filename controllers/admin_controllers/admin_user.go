package admin_controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/internal/database"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/internal/websocket"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/services"
	"github.com/leleo886/lopic/services/admin_services"
)

type UserController struct {
	naUserService *services.UserService
	userService   *admin_services.UserService
	hub           *websocket.Hub
}

func NewUserController(userService *services.UserService, adminUserService *admin_services.UserService, hub *websocket.Hub) *UserController {
	return &UserController{naUserService: userService, userService: adminUserService, hub: hub}
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取所有用户，支持分页
// @Tags 用户管理员
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为10"
// @Param searchkey query string false "搜索关键词"
// @Param orderby query string false "排序字段，默认为created_at"
// @Param order query string false "排序顺序，asc或desc，默认为desc"
// @Security ApiKeyAuth
// @Success 200 {object} success.DataResponse{data=admin_services.GetUsersResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/users [get]
func (h *UserController) GetUsers(c *gin.Context) {
	// 获取分页参数
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

	users, err := h.userService.GetUsers(page, pageSize, offset, searchkey, orderby, order)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Users retrieved successfully", users))
}

// GetUser 获取单个用户
// @Summary 获取单个用户
// @Description 根据ID获取用户信息
// @Tags 用户管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} success.DataResponse{data=models.User}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/users/{id} [get]
func (h *UserController) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("User retrieved successfully", user))
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理员
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body admin_services.UserRequest true "用户信息"
// @Success 201 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/users [post]
func (h *UserController) CreateUser(c *gin.Context) {
	var req admin_services.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if !isValidRole(req.Role) {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if req.Role == "admin" {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrOneAdmin)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 创建用户
	err := h.userService.CreateUser(req)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	// tx.Commit() // 提交事务
	c.JSON(http.StatusCreated, success.NewSuccessResponse("User created successfully"))
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 根据ID更新用户信息
// @Tags 用户管理员
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Param user body admin_services.UserRequest true "用户信息"
// @Success 200 {object} success.DataResponse{data=models.User}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/users/{id} [put]
func (h *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	var req admin_services.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	if !isValidRole(req.Role) {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	user, err := h.userService.UpdateUser(id, currentUserID.(uint), req)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("User updated successfully", user))
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据ID删除用户且会删除用户所有关联的相册和图片
// @Tags 用户管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/users/{id} [delete]
func (h *UserController) DeleteUser(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 立即返回响应，告知客户端删除已开始
	c.JSON(http.StatusOK, success.NewSuccessResponse("User deletion started"))

	err = h.userService.DeleteUser(id)
	if err != nil {
		_, errorResponse := cerrors.NewErrorResponse(err)
		log.Errorf("Failed to delete user %d: %v", id, err)
		h.hub.BroadcastToUser(currentUserID.(uint), "delete_user_error", map[string]interface{}{
			"message": "Failed to delete user, some images may not be deleted",
			"error":   errorResponse.Message,
			"code":    errorResponse.Code,
		})
	} else {
		h.hub.BroadcastToUser(currentUserID.(uint), "delete_user_success", map[string]interface{}{
			"message": "User deleted successfully",
		})
	}
}

func isValidRole(roleName string) bool {
	var role models.Role
	result := database.GetDB().Where("name = ?", roleName).First(&role)
	return result.RowsAffected > 0
}

// GetUserImagesTagsCloud 获取用户图片标签云
// @Summary 获取用户图片标签云
// @Description 获取指定用户所有图片的标签云（标签及出现次数）
// @Tags 用户管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} success.DataResponse{data=[]services.TagCloudItem}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/users/{id}/tags-cloud [get]
func (h *UserController) GetUserImagesTagsCloud(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}
	tagsCloud, err := h.naUserService.GetImagesTagsCloud(uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("User tags cloud retrieved successfully", tagsCloud))
}

// GetAllImagesTagsCloud 获取所有用户图片标签云
// @Summary 获取所有用户图片标签云
// @Description 获取所有用户所有图片的标签云（标签及出现次数）
// @Tags 用户管理员
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} success.DataResponse{data=[]services.TagCloudItem}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/users/tags-cloud [get]
func (h *UserController) GetAllImagesTagsCloud(c *gin.Context) {
	tagsCloud, err := h.userService.GetAllImagesTagsCloud()
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("All users tags cloud retrieved successfully", tagsCloud))
}
