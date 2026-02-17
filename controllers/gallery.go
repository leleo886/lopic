package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/services"
)

type GalleryController struct {
	galleryService *services.GalleryService
	galleryConfig  *models.GalleryConfig
}

func NewGalleryController(galleryService *services.GalleryService, galleryConfig *models.GalleryConfig) *GalleryController {
	return &GalleryController{galleryService: galleryService, galleryConfig: galleryConfig}
}

// GetGallerys 获取所有画廊相册
// @Summary 获取所有画廊相册
// @Description 获取所有画廊相册
// @Tags 画廊
// @Produce json
// @Param user_name path string true "用户名"
// @Success 200 {object} success.DataResponse{data=services.GetAlbumsResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/gallery/albums/{user_name} [get]
func (h *GalleryController) GetGallerys(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}

	gallery, err := h.galleryService.GetGallery(currentUserID.(uint))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("get gallerys success", gallery))
}

// GetGalleryImages 获取画廊相册图片
// @Summary 获取画廊相册图片
// @Description 获取画廊相册图片
// @Tags 画廊
// @Produce json
// @Param user_name path string true "用户名"
// @Param album_id path int true "相册ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} success.DataResponse{data=services.GetAlbumImagesResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/gallery/images/{user_name}/{album_id} [get]
func (h *GalleryController) GetGalleryImages(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}
	albumID, err := strconv.Atoi(c.Param("album_id"))
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

	images, err := h.galleryService.GetGalleryImages(currentUserID.(uint), uint(albumID), page, pageSize, offset)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("get images success", images))
}

// SearchGalleryImages 搜索画廊相册图片
// @Summary 搜索画廊相册图片
// @Description 搜索画廊相册图片
// @Tags 画廊
// @Produce json
// @Param user_name path string true "用户名"
// @Param query query string true "搜索查询"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} success.DataResponse{data=services.GetGalleryImagesResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/gallery/search/{user_name} [get]
func (h *GalleryController) SearchGalleryImages(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
		c.JSON(statusCode, errorResponse)
		return
	}
	query := c.DefaultQuery("query", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	images, err := h.galleryService.SearchGalleryImages(currentUserID.(uint), query, page, pageSize, offset)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("search images success", images))
}

// GetGalleryConfig 获取画廊配置
// @Summary 获取画廊配置
// @Description 获取画廊配置数据
// @Tags 画廊
// @Produce json
// @Success 200 {object} success.DataResponse{data=models.GalleryConfig}
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/gallery/config [get]
func (h *GalleryController) GetGalleryConfig(c *gin.Context) {
	c.JSON(http.StatusOK, success.NewDataResponse("get gallery config success", h.galleryConfig))
}
