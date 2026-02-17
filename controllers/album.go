package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/services"
)

// AlbumController
type AlbumController struct {
	albumService *services.AlbumService
}

// NewAlbumController
func NewAlbumController(albumService *services.AlbumService) *AlbumController {
	return &AlbumController{albumService: albumService}
}

type AlbumRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=100"`
	Description    string `json:"description" binding:"max=500"`
	GalleryEnabled bool   `json:"gallery_enabled"`
	SerialNumber   int    `json:"serial_number"`
}

// CreateAlbum 创建相册
// @Summary 创建相册
// @Description 创建新的相册
// @Tags 相册管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param album body AlbumRequest true "相册信息"
// @Success 200 {object} success.DataResponse{data=services.AlbumResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/albums [post]
func (h *AlbumController) CreateAlbum(c *gin.Context) {
	var req AlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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

	albumResponse, err := h.albumService.CreateAlbum(req.Name, req.Description, req.GalleryEnabled, currentUserID.(uint), req.SerialNumber)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Album created successfully", albumResponse))
}

// GetAlbums 获取相册列表
// @Summary 获取相册列表
// @Description 获取当前用户的相册列表
// @Tags 相册管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} success.DataResponse{data=services.GetAlbumsResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/albums [get]
func (h *AlbumController) GetAlbums(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
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

	offset := (page - 1) * pageSize

	albumResponses, err := h.albumService.GetAlbums(currentUserID.(uint), page, pageSize, offset)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Albums retrieved successfully", albumResponses))
}

// GetAlbum 获取相册详情
// @Summary 获取相册详情
// @Description 获取指定相册的详细信息
// @Tags 相册管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "相册ID"
// @Success 200 {object} success.DataResponse{data=services.AlbumResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/albums/{id} [get]
func (h *AlbumController) GetAlbum(c *gin.Context) {
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

	albumResponse, err := h.albumService.GetAlbum(uint(id), currentUserID.(uint))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Album retrieved successfully", albumResponse))
}

// UpdateAlbum 更新相册
// @Summary 更新相册
// @Description 更新指定相册的信息
// @Tags 相册管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "相册ID"
// @Param album body AlbumRequest true "相册信息"
// @Success 200 {object} success.DataResponse{data=services.AlbumResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/albums/{id} [put]
func (h *AlbumController) UpdateAlbum(c *gin.Context) {
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

	var req AlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	albumResponse, err := h.albumService.UpdateAlbum(uint(id), currentUserID.(uint), req.Name, req.Description, req.GalleryEnabled, req.SerialNumber)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Album updated successfully", albumResponse))
}

// DeleteAlbum 删除相册
// @Summary 删除相册，只删除相册与图片的关联信息，不删除图片
// @Description 删除指定的相册（不会删除图片）
// @Tags 相册管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "相册ID"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/albums/{id} [delete]
func (h *AlbumController) DeleteAlbum(c *gin.Context) {
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

	err = h.albumService.DeleteAlbum(uint(id), currentUserID.(uint))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Delete album success"))
}

// GetAlbumImages 获取相册中的图片列表
// @Summary 获取相册中的图片
// @Description 获取指定相册中的所有图片
// @Tags 相册管理
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "相册ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} success.DataResponse{data=services.GetAlbumImagesResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/albums/{id}/images [get]
func (h *AlbumController) GetAlbumImages(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	imageResponses, err := h.albumService.GetAlbumImages(uint(id), currentUserID.(uint), page, pageSize, offset)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Album images retrieved successfully", imageResponses))
}

// GetNotInAnyAlbum 获取不在任何相册中的图片列表
// @Summary 获取不在任何相册中的图片
// @Description 获取所有不在任何相册中的图片
// @Tags 相册管理
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} success.DataResponse{data=services.GetAlbumImagesResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/albums/images/not-in-any [get]
func (h *AlbumController) GetNotInAnyAlbum(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
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

	offset := (page - 1) * pageSize

	imageResponses, err := h.albumService.GetNotInAnyAlbum(currentUserID.(uint), page, pageSize, offset)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Images not in any album retrieved successfully", imageResponses))

}
