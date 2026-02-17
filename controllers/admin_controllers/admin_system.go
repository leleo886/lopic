package admin_controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/database"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/mail"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/models"
)

type SystemController struct {
	mailService   *mail.MailService
	generalConfig *models.GeneralConfig
	galleryConfig *models.GalleryConfig
}

func NewSystemController(mailService *mail.MailService, generalConfig *models.GeneralConfig, galleryConfig *models.GalleryConfig) *SystemController {
	return &SystemController{mailService: mailService, generalConfig: generalConfig, galleryConfig: galleryConfig}
}

// GetSystemInfo 获取系统信息
// @Summary 获取系统信息
// @Description 获取系统信息
// @Tags 系统
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} success.DataResponse{data=models.SystemSettings}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/system/info [get]
func (s *SystemController) GetSystemInfo(c *gin.Context) {
	systemSettings, err := config.LoadSystemSettingsFromDatabase(database.GetDB())
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 禁止明文密码返回
	systemSettings.Mail.SMTP.Password = ""

	c.JSON(http.StatusOK, success.NewDataResponse("System info retrieved successfully", systemSettings))
}

// UpdateSystemInfo 更新系统信息
// @Summary 更新系统信息
// @Description 更新系统信息
// @Tags 系统
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param systemSettings body models.SystemSettings true "系统信息"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/admin/system/info [put]
func (s *SystemController) UpdateSystemInfo(c *gin.Context) {
	var req models.SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if req.Mail.SMTP.Password == "" {
		req.Mail.SMTP.Password = s.mailService.GetSMTPPWD()
	}
	if req.General.MaxTags < 0  {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	err := config.ImportSystemSettingsToDatabase(database.GetDB(), req)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 更新系统配置
	*s.generalConfig = req.General
	*s.galleryConfig = req.Gallery

	// 为了避免热更新问题，手动更新邮件服务配置
	s.mailService.UpdateConfig(&req.Mail)

	c.JSON(http.StatusOK, success.NewSuccessResponse("System info updated successfully"))
}
