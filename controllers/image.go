package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/internal/websocket"
	"github.com/leleo886/lopic/services"
)

type ImageController struct {
	imageService *services.ImageService
	hub          *websocket.Hub
}

func NewImageController(imageService *services.ImageService, hub *websocket.Hub) *ImageController {
	return &ImageController{
		imageService: imageService,
		hub:          hub,
	}
}

type UploadRequest struct {
	Tags     []string `form:"tags" binding:"omitempty"`
	AlbumIDs []uint   `form:"album_ids" binding:"omitempty"`
}

type UpdateRequest struct {
	IDs          []uint   `json:"ids" binding:"required"`
	OriginalName string   `json:"original_name" binding:"omitempty"`
	Tags         []string `json:"tags" binding:"omitempty"`
}

type UpdateResponse struct {
	SuccessIDs map[uint]*services.ImageResponse `json:"success_ids"`
	ErrorIDs   map[uint]string                  `json:"error_ids"`
}

type AddOrDelImageToAlbumRequset struct {
	IDs     []uint `json:"ids" binding:"required"`
	AlbumID uint   `json:"album_id" binding:"required"`
}

type AddOrDelImageToAlbumResponse struct {
	SuccessIDs map[uint]string `json:"success_ids"`
	ErrorIDs   map[uint]string `json:"error_ids"`
}

// UploadImage 批量上传图片
// @Summary 批量上传图片
// @Description 上传图片文件
// @Tags 图片管理
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData []file true "图片文件"
// @Param tags formData []string false "标签列表"
// @Param album_ids formData []int false "相册ID列表"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images/upload [post]
func (h *ImageController) UploadImage(c *gin.Context) {
	var req UploadRequest
	if err := c.ShouldBind(&req); err != nil {
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

	form, err := c.MultipartForm()
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}
	
	// 立即返回响应，告知客户端上传已开始
	c.JSON(http.StatusOK, success.NewSuccessResponse("Upload started"))

	// 获取所有名为 "file" 的上传文件
	files := form.File["file"]
	if len(files) == 0 {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	err = h.imageService.UploadImageLimitCheck(currentUserID.(uint), files)
	if err != nil {
		_, errorResponse := cerrors.NewErrorResponse(err)
		h.hub.BroadcastToUser(currentUserID.(uint), "upload_processing_error", map[string]interface{}{
			"message": "file not allowed",
			"error":   errorResponse.Message,
			"code":    errorResponse.Code,
		})
		return
	}

	uniqueTags := make(map[string]bool)
	for _, tag := range req.Tags {
		uniqueTags[tag] = true
	}
	req.Tags = nil
	for tag := range uniqueTags {
		req.Tags = append(req.Tags, tag)
	}

	// 在后台处理文件上传
	go func() {
		userID := currentUserID.(uint)

		// 发送内部处理开始的消息
		h.hub.BroadcastToUser(userID, "upload_processing_start", map[string]interface{}{
			"message":    "Start processing images",
			"file_count": len(files),
		})

		err = h.imageService.UploadImage(
			userID,
			req.AlbumIDs,
			req.Tags,
			files,
		)

		if err != nil {
			_, errorResponse := cerrors.NewErrorResponse(err)
			// 发送内部处理错误的消息
			h.hub.BroadcastToUser(userID, "upload_processing_error", map[string]interface{}{
				"message": "Image processing failed",
				"error":   errorResponse.Message,
				"code":    errorResponse.Code,
			})
			log.Errorf("Failed to process upload: %v", err)
		} else {
			// 发送内部处理完成的消息
			h.hub.BroadcastToUser(userID, "upload_processing_complete", map[string]interface{}{
				"message":    "Image processing completed",
				"file_count": len(files),
			})
		}
	}()
}

// GetImages 获取图片列表
// @Summary 获取图片列表
// @Description 获取当前用户的图片列表
// @Tags 图片管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} success.DataResponse{data=services.GetImagesResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images [get]
func (h *ImageController) GetImages(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	imagesResponse, err := h.imageService.GetImages(currentUserID.(uint), page, pageSize)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Images retrieved successfully", imagesResponse))
}

// GetImage 获取单张图片详情
// @Summary 获取图片详情
// @Description 获取指定图片的详细信息
// @Tags 图片管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "图片ID"
// @Success 200 {object} success.DataResponse{data=services.ImageResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images/{id} [get]
func (h *ImageController) GetImage(c *gin.Context) {
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

	imageResponse, err := h.imageService.GetImage(currentUserID.(uint), uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Image retrieved successfully", imageResponse))
}

// UpdateImage 更新图片
// @Summary 更新图片
// @Description 更新指定图片的信息
// @Tags 图片管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param image body UpdateRequest true "图片信息"
// @Success 200 {object} success.DataResponse{data=UpdateResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images [put]
func (h *ImageController) UpdateImage(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	fmt.Println(req)

	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	var ErrorIDs map[uint]string
	var SuccessIDs map[uint]*services.ImageResponse

	uniqueTags := make(map[string]bool)
	for _, tag := range req.Tags {
		uniqueTags[tag] = true
	}
	req.Tags = nil
	for tag := range uniqueTags {
		req.Tags = append(req.Tags, tag)
	}

	for _, id := range req.IDs {
		imageResponse, err := h.imageService.UpdateImage(currentUserID.(uint), id, req.OriginalName, req.Tags)
		if err != nil {
			if ErrorIDs == nil {
				ErrorIDs = make(map[uint]string)
			}
			ErrorIDs[id] = err.Error()
		} else {
			if SuccessIDs == nil {
				SuccessIDs = make(map[uint]*services.ImageResponse)
			}
			SuccessIDs[id] = imageResponse
		}
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Images updated successfully", &UpdateResponse{
		SuccessIDs: SuccessIDs,
		ErrorIDs:   ErrorIDs,
	}))
}

// DeleteImage 删除图片
// @Summary 删除图片
// @Description 删除指定的图片
// @Tags 图片管理
// @Produce json
// @Security ApiKeyAuth
// @Param ids body []uint true "图片ID"
// @Success 200 {object} success.NewSuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images [delete]
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
		err := h.imageService.DeleteImage(currentUserID.(uint), id)
		if err != nil {
			log.Errorf("Failed to delete image %d: %v", id, err)
			isExistError = true
			ExistError = err
		} 
	}

	if isExistError {
		_, errorResponse := cerrors.NewErrorResponse(ExistError)
		h.hub.BroadcastToUser(currentUserID.(uint), "delete_exist_error", map[string]interface{}{
			"message":    "Exist error, some images may not be deleted",
			"error":      errorResponse.Message,
			"code":       errorResponse.Code,
		})
	} else {
		h.hub.BroadcastToUser(currentUserID.(uint), "delete_success", map[string]interface{}{
			"message":    "Images deleted successfully",
		})
	}
}

// AddImageToAlbum 批量添加图片到相册
// @Summary 添加图片到相册
// @Description 将图片添加到指定相册
// @Tags 图片管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param req body AddOrDelImageToAlbumRequset true "添加图片到相册请求"
// @Success 200 {object} success.DataResponse{data=AddOrDelImageToAlbumResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images/albums [post]
func (h *ImageController) AddImageToAlbum(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	var req AddOrDelImageToAlbumRequset
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

	var ErrorIDs map[uint]string
	var SuccessIDs map[uint]string
	for _, id := range req.IDs {
		err := h.imageService.AddImageToAlbum(currentUserID.(uint), id, req.AlbumID)
		if err != nil {
			if errors.Is(err, cerrors.ErrAlbumNotFound) {
				statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrAlbumNotFound)
				c.JSON(statusCode, errorResponse)
				return
			} else {
				if ErrorIDs == nil {
					ErrorIDs = make(map[uint]string)
				}
				ErrorIDs[id] = err.Error()
			}
		} else {
			if SuccessIDs == nil {
				SuccessIDs = make(map[uint]string)
			}
			SuccessIDs[id] = "Add to album success"
		}
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Images added to album successfully", &AddOrDelImageToAlbumResponse{
		SuccessIDs: SuccessIDs,
		ErrorIDs:   ErrorIDs,
	}))
}

// RemoveImageFromAlbum 将图片从指定相册移除
// @Summary 从相册移除图片
// @Description 将图片从指定相册移除
// @Tags 图片管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param req body AddOrDelImageToAlbumRequset true "从相册移除图片请求"
// @Success 200 {object} success.DataResponse{data=AddOrDelImageToAlbumResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images/albums [delete]
func (h *ImageController) RemoveImageFromAlbum(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	var req AddOrDelImageToAlbumRequset
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

	var ErrorIDs map[uint]string
	var SuccessIDs map[uint]string

	for _, id := range req.IDs {
		err := h.imageService.RemoveImageFromAlbum(currentUserID.(uint), id, req.AlbumID)
		if err != nil {
			if errors.Is(err, cerrors.ErrAlbumNotFound) {
				statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrAlbumNotFound)
				c.JSON(statusCode, errorResponse)
				return
			} else {
				if ErrorIDs == nil {
					ErrorIDs = make(map[uint]string)
				}
				ErrorIDs[id] = err.Error()
			}
		} else {
			if SuccessIDs == nil {
				SuccessIDs = make(map[uint]string)
			}
			SuccessIDs[id] = "Remove from album success"
		}
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Images removed from album successfully", &AddOrDelImageToAlbumResponse{
		SuccessIDs: SuccessIDs,
		ErrorIDs:   ErrorIDs,
	}))
}

// SearchImagesByTagsOrTitle 标签或标题模糊搜索图片8
// @Summary 标签或标题模糊搜索图片
// @Description 标签、标题只提供其一可针对提供值的字段进行模糊搜索，若都提供则对两个值进行模糊搜索
// @Tags 图片管理
// @Produce json
// @Security ApiKeyAuth
// @Param search_key query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} success.DataResponse{data=services.GetImagesResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/images/search [get]
func (h *ImageController) SearchImagesByTagsOrTitle(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	searchKey := c.Query("search_key")
	if searchKey == "" {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	imagesResponse, err := h.imageService.SearchImagesByTagsOrTitle(currentUserID.(uint), searchKey, page, pageSize)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Images searched successfully", imagesResponse))
}
