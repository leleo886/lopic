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

type StorageController struct {
	storageService *admin_services.StorageService
}

func NewStorageController(storageService *admin_services.StorageService) *StorageController {
	return &StorageController{storageService: storageService}
}

// GetStorages 获取存储配置列表
// @Summary 获取存储配置列表
// @Description 获取所有存储配置，支持分页
// @Tags 存储管理员
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param searchkey query string false "搜索关键词"
// @Param orderby query string false "排序字段" default(created_at)
// @Param order query string false "排序方向" default(desc)
// @Success 200 {object} success.DataResponse{data=[]models.Storage}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/storages [get]
func (h *StorageController) GetStorages(c *gin.Context) {
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

	storages, err := h.storageService.GetStorages(page, pageSize, offset, searchkey, orderby, order)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	// 移除存储配置中的密码信息
	cleanStorages := removeStoragesPassword(*storages)
	c.JSON(http.StatusOK, success.NewDataResponse("Storages retrieved successfully", cleanStorages))
}

// GetStorage 获取存储配置详情
// @Summary 获取存储配置详情
// @Description 根据ID获取存储配置详情
// @Tags 存储管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "存储配置ID"
// @Success 200 {object} success.DataResponse{data=models.Storage}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/storages/{id} [get]
func (h *StorageController) GetStorage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	storage, err := h.storageService.GetStorage(id)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	// 移除存储配置中的密码信息
	cleanStorage := removeStoragePassword(storage)
	c.JSON(http.StatusOK, success.NewDataResponse("Storage retrieved successfully", cleanStorage))
}

// GetStorageByName 根据名称获取存储配置详情
// @Summary 根据名称获取存储配置详情
// @Description 根据存储名称获取存储配置详情
// @Tags 存储管理员
// @Produce json
// @Security ApiKeyAuth
// @Param name path string true "存储配置名称"
// @Success 200 {object} success.DataResponse{data=models.Storage}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/storages/name/{name} [get]
func (h *StorageController) GetStorageByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	storage, err := h.storageService.GetStorageByName(name)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	// 移除存储配置中的密码信息
	cleanStorage := removeStoragePassword(storage)
	c.JSON(http.StatusOK, success.NewDataResponse("Storage retrieved successfully", cleanStorage))
}

// CreateStorage 创建存储配置
// @Summary 创建存储配置
// @Description 创建一个新存储配置
// @Tags 存储管理员
// @Produce json
// @Security ApiKeyAuth
// @Param storage body admin_services.StorageRequest true "存储配置信息"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/storages [post]
func (h *StorageController) CreateStorage(c *gin.Context) {
	var storage admin_services.StorageRequest
	if err := c.ShouldBindJSON(&storage); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if err := storageRequestCheck(&storage); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	err := h.storageService.CreateStorage(&storage)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Storage created successfully"))
}

// UpdateStorage 更新存储配置
// @Summary 更新存储配置
// @Description 根据ID更新存储配置信息
// @Tags 存储管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "存储配置ID"
// @Param storage body admin_services.StorageRequest true "存储配置信息"
// @Success 200 {object} success.DataResponse{data=admin_services.StorageRequest}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/storages/{id} [put]
func (h *StorageController) UpdateStorage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	var storage admin_services.StorageRequest
	if err := c.ShouldBindJSON(&storage); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if err := storageRequestCheck(&storage); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	err = h.storageService.UpdateStorage(uint(id), &storage)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	// 移除存储配置中的密码信息，替换为固定占位符
	if storage.Type == "webdav" {
		// 使用固定占位符
		storage.Config.Password = "[SET]"
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Storage updated successfully", storage))

}

// TestStorageConnection 测试存储连接
// @Summary 测试存储连接
// @Description 测试存储配置的连接状态
// @Tags 存储管理员
// @Produce json
// @Security ApiKeyAuth
// @Param storage body admin_services.StorageRequest true "存储配置信息"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/storages/test [post]
func (h *StorageController) TestStorageConnection(c *gin.Context) {
	var storage admin_services.StorageRequest
	if err := c.ShouldBindJSON(&storage); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if err := storageRequestCheck(&storage); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	err := h.storageService.TestStorageConnection(&storage)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Storage connection test successful"))
}

// DeleteStorage 删除存储配置
// @Summary 删除存储配置
// @Description 根据ID删除存储配置
// @Tags 存储管理员
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "存储配置ID"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/storages/{id} [delete]
func (h *StorageController) DeleteStorage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	err = h.storageService.DeleteStorage(uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewSuccessResponse("Storage deleted successfully"))
}

func storageRequestCheck(storage *admin_services.StorageRequest) error {
	if storage.Name == "" {
		return cerrors.ErrBadRequest
	}
	if storage.Type != "local" && storage.Type != "webdav" {
		return cerrors.ErrBadRequest
	}
	if storage.Type == "webdav" {
		if storage.Config.BaseURL == "" || storage.Config.Username == "" || storage.Config.StaticURL == "" || storage.Config.BasePath == "" {
			return cerrors.ErrBadRequest
		}
		// 移除密码的必填验证，允许留空保持原密码
	}
	if storage.Type == "local" {
		// local 类型的存储只能有一个，不能再创建
		return cerrors.ErrBadRequest
	}
	return nil
}

// removeStoragePassword 移除存储配置中的密码信息，替换为密码设置状态
func removeStoragePassword(storage *models.Storage) *models.Storage {
	if storage.Type == "webdav" {
		// 创建一个新的 StorageConfig 副本，移除密码，添加密码设置状态
		configCopy := storage.Config
		// 移除密码，替换为固定占位符
		configCopy.Password = "[SET]"
		// 创建一个新的 Storage 副本
		storageCopy := *storage
		storageCopy.Config = configCopy
		return &storageCopy
	}
	return storage
}

// removeStoragesPassword 移除多个存储配置中的密码信息
func removeStoragesPassword(storages []models.Storage) []models.Storage {
	result := make([]models.Storage, len(storages))
	for i, storage := range storages {
		storageCopy := *removeStoragePassword(&storage)
		result[i] = storageCopy
	}
	return result
}
