package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/services"
	cerrors "github.com/leleo886/lopic/internal/error"
	"net/http"
)

type UserController struct {
	userService   *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

// UserRequest 用户管理请求结构体
type UserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

// GetMe 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的信息
// @Tags 用户查询
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.User
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/users/me [get]
func (h *UserController) GetMe(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	user, err := h.userService.GetMe(currentUserID.(uint))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUserNotFound)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateMe 更新当前用户信息
// @Summary 更新当前用户信息
// @Description 更新当前登录用户的信息
// @Tags 用户查询
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body UserRequest true "用户信息"
// @Success 200 {object} models.User
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/users/me [put]
func (h *UserController) UpdateMe(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 获取当前登录用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 更新用户信息
	user, err := h.userService.UpdateMe(currentUserID.(uint), req.Username, req.Password)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, user)
}

// GetStorageUsage 获取当前用户存储空间使用情况
// @Summary 获取当前用户存储空间使用情况
// @Description 获取当前登录用户的存储空间使用情况
// @Tags 用户查询
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} services.StorageUsage
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/users/me/storage [get]
func (h *UserController) GetStorageUsage(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}
	storageUsage, err := h.userService.GetStorageUsage(currentUserID.(uint))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, storageUsage)
}

// GetImagesTagsCloud 获取用户图片标签云
// @Summary 获取用户图片标签云
// @Description 获取当前登录用户所有图片的标签云（标签及出现次数）
// @Tags 用户查询
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} []services.TagCloudItem
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/users/me/tags-cloud [get]
func (h *UserController) GetImagesTagsCloud(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}
	tagsCloud, err := h.userService.GetImagesTagsCloud(currentUserID.(uint))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, tagsCloud)
}

