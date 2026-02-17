package admin_controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/controllers"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/services/admin_services"
)

type AlbumController struct {
	albumService *admin_services.AlbumService
}

func NewAlbumController(albumService *admin_services.AlbumService) *AlbumController {
	return &AlbumController{albumService: albumService}
}

// GetAllAlbums 获取所有相册列表
// @Summary 获取所有相册列表
// @Description 获取所有相册，支持分页
// @Tags 相册管理员
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为10"
// @Param searchkey query string false "搜索关键词"
// @Param orderby query string false "排序字段，默认为created_at"
// @Param order query string false "排序方向，asc或desc，默认为desc"
// @Security ApiKeyAuth
// @Success 200 {object} success.DataResponse{data=services.GetAlbumsResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/albums [get]
func (h *AlbumController) GetAllAlbums(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	searchkey := c.DefaultQuery("searchkey", "")
	orderby := c.DefaultQuery("orderby", "created_at")
	order := c.DefaultQuery("order", "desc")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	if order != "asc" {
		order = "desc"
	}
	orderby = strings.TrimSpace(orderby)
	order = strings.TrimSpace(order)

	albums, err := h.albumService.GetAllAlbums(page, pageSize, offset, searchkey, orderby, order)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Albums retrieved successfully", albums))
}

// GetAlbum 获取相册详情
// @Summary 获取相册详情
// @Description 获取指定相册的详细信息
// @Tags 相册管理员
// @Produce json
// @Param id path int true "相册ID"
// @Security ApiKeyAuth
// @Success 200 {object} success.DataResponse{data=services.AlbumResponse}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/albums/{id} [get]
func (h *AlbumController) GetAlbum(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	album, err := h.albumService.GetAlbum(uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}
	c.JSON(http.StatusOK, success.NewDataResponse("Album retrieved successfully", album))
}

// DeleteAlbum 删除相册
// @Summary 删除相册，只删除相册与图片的关联信息，不删除图片
// @Description 删除指定相册
// @Tags 相册管理员
// @Produce json
// @Param ids body []uint true "相册ID列表"
// @Security ApiKeyAuth
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/albums [delete]
func (h *AlbumController) DeleteAlbum(c *gin.Context) {
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

	var ErrorIDs map[uint]string
	var SuccessIDs map[uint]string
	for _, id := range ids {
		err := h.albumService.DeleteAlbum(id)
		if err != nil {
			if ErrorIDs == nil {
				ErrorIDs = make(map[uint]string)
			}
			ErrorIDs[id] = err.Error()
		} else {
			if SuccessIDs == nil {
				SuccessIDs = make(map[uint]string)
			}
			SuccessIDs[id] = "Delete success"
		}
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Albums deleted successfully", &controllers.AddOrDelImageToAlbumResponse{
		SuccessIDs: SuccessIDs,
		ErrorIDs:   ErrorIDs,
	}))
}
