package admin_controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/controllers"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/internal/websocket"
	"github.com/leleo886/lopic/services/admin_services"
)

type ImageController struct {
	imageService *admin_services.ImageService
	hub          *websocket.Hub
}

func NewImageController(imageService *admin_services.ImageService, hub *websocket.Hub) *ImageController {
	return &ImageController{imageService: imageService, hub: hub}
}

// GetAllImages 获取所有图片列表
// @Summary 获取所有图片列表
// @Description 获取所有图片，支持分页和字段值查询
// @Tags 图片管理员
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为10"
// @Param searchkey query string false "搜索关键词"
// @Param field query string false "查询字段名"
// @Param value query string false "查询字段值"
// @Param orderby query string false "排序字段，默认为created_at"
// @Param order query string false "排序方向，asc或desc，默认为desc"
// @Security ApiKeyAuth
// @Success 200 {object} services.GetImagesResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/images [get]
func (h *ImageController) GetAllImages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	searchkey := c.DefaultQuery("searchkey", "")
	field := c.DefaultQuery("field", "")
	value := c.DefaultQuery("value", "")
	orderby := c.DefaultQuery("orderby", "created_at")
	order := c.DefaultQuery("order", "desc")

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

	images, err := h.imageService.GetAllImages(page, pageSize, offset, searchkey, field, value, orderby, order)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, images)
}

// GetImage 获取单张图片
// @Summary 获取单张图片
// @Description 获取指定ID的图片详情
// @Tags 图片管理员
// @Produce json
// @Param id path int true "图片ID"
// @Security ApiKeyAuth
// @Success 200 {object} services.ImageResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/images/{id} [get]
func (h *ImageController) GetImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	imageResponse, err := h.imageService.GetImage(uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, imageResponse)
}

// DeleteImage 删除图片
// @Summary 删除图片
// @Description 删除指定ID的图片
// @Tags 图片管理员
// @Produce json
// @Param ids body []uint true "图片ID"
// @Security ApiKeyAuth
// @Success 200 {object} success.NewSuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/images [delete]
func (h *ImageController) DeleteImage(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}
	// remove duplicates
	uniqueIDs := make(map[uint]bool)
	for _, id := range ids {
		uniqueIDs[id] = true
	}
	ids = nil
	for id := range uniqueIDs {
		ids = append(ids, id)
	}

	// 立即返回响应，告知客户端删除已开始
	c.JSON(http.StatusOK, success.NewSuccessResponse("Images deleted started"))

	// delete images
	isExistError := false
	var ExistError error
	for _, id := range ids {
		err := h.imageService.DeleteImage(id)
		if err != nil {
			log.Errorf("Failed to delete image %d: %v", id, err)
			isExistError = true
			ExistError = err
		}
	}

	if isExistError {
		_, errorResponse := cerrors.NewErrorResponse(ExistError)
		h.hub.BroadcastToUser(currentUserID.(uint), "delete_exist_error", map[string]interface{}{
			"message": "Exist error, some images may not be deleted",
			"error":   errorResponse.Message,
			"code":    errorResponse.Code,
		})
	} else {
		h.hub.BroadcastToUser(currentUserID.(uint), "delete_success", map[string]interface{}{
			"message": "Images deleted successfully",
		})
	}
}

// UpdateImageStorage 更新图片存储名称
// @Summary 更新图片存储名称
// @Description 更新指定图片的存储名称
// @Tags 图片管理员
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body UpdateStorageRequest true "更新存储名称请求"
// @Success 200 {object} controllers.AddOrDelImageToAlbumResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/images/storagename [put]
func (h *ImageController) UpdateImageStorage(c *gin.Context) {
	var req UpdateStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	// remove duplicates
	uniqueIDs := make(map[uint]bool)
	for _, id := range req.IDs {
		uniqueIDs[id] = true
	}
	req.IDs = nil
	for id := range uniqueIDs {
		req.IDs = append(req.IDs, id)
	}

	// update images storage
	var ErrorIDs map[uint]string
	var SuccessIDs map[uint]string
	for _, id := range req.IDs {
		err := h.imageService.UpdateImageStorage(id, req.StorageName)
		if err != nil {
			if ErrorIDs == nil {
				ErrorIDs = make(map[uint]string)
			}
			ErrorIDs[id] = err.Error()
		} else {
			if SuccessIDs == nil {
				SuccessIDs = make(map[uint]string)
			}
			SuccessIDs[id] = "Update storage success"
		}
	}

	c.JSON(http.StatusOK, &controllers.AddOrDelImageToAlbumResponse{
		SuccessIDs: SuccessIDs,
		ErrorIDs:   ErrorIDs,
	})

}

// UpdateStorageRequest 更新存储名称请求
type UpdateStorageRequest struct {
	IDs         []uint `json:"ids" binding:"required"`
	StorageName string `json:"storage_name" binding:"required"`
}
