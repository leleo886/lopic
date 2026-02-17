package admin_controllers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/services/admin_services"
)

// BackupController 备份控制器
type BackupController struct {
	backupService *admin_services.BackupService
}

// NewBackupController 创建备份控制器实例
func NewBackupController(backupService *admin_services.BackupService) *BackupController {
	return &BackupController{
		backupService: backupService,
	}
}

// CreateBackup 创建备份
// @Summary 创建备份
// @Description 创建系统全量备份
// @Tags backup
// @Accept json
// @Produce json
// @Success 200 {object} success.DataResponse{data=models.BackupTask}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup [post]
func (h *BackupController) CreateBackup(c *gin.Context) {
	task, err := h.backupService.CreateBackup()
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Backup created successfully", task))
}

// GetBackupList 获取备份列表
// @Summary 获取备份列表
// @Description 获取所有备份任务列表
// @Tags backup
// @Accept json
// @Produce json
// @Success 200 {object} success.DataResponse{data=[]models.BackupTask}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup/list [get]
func (h *BackupController) GetBackupList(c *gin.Context) {
	tasks, err := h.backupService.GetBackupList()
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Backup list retrieved successfully", tasks))
}

// GetRestoreRecords 获取恢复记录
// @Summary 获取恢复记录
// @Description 获取所有恢复任务记录
// @Tags backup
// @Accept json
// @Produce json
// @Success 200 {object} success.DataResponse{data=[]models.RestoreTask}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup/restore/list [get]
func (h *BackupController) GetRestoreRecords(c *gin.Context) {
	records, err := h.backupService.GetRestoreRecords()
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Restore records retrieved successfully", records))
}

// DeleteBackup 删除备份
// @Summary 删除备份
// @Description 删除指定的备份
// @Tags backup
// @Accept json
// @Produce json
// @Param id path int true "备份ID"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup/{id} [delete]
func (h *BackupController) DeleteBackup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if err := h.backupService.DeleteBackup(uint(id)); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("备份删除成功"))
}

// RestoreBackup 恢复备份
// @Summary 恢复备份
// @Description 从指定备份恢复系统
// @Tags backup
// @Accept json
// @Produce json
// @Param id path int true "备份ID"
// @Success 200 {object} success.DataResponse{data=models.RestoreTask}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup/restore/{id} [post]
func (h *BackupController) RestoreBackup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	task, err := h.backupService.RestoreBackup(uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Backup restore task created successfully", task))
}

// DeleteRestoreTask 删除恢复任务
// @Summary 删除恢复任务
// @Description 删除指定的恢复任务记录
// @Tags backup
// @Accept json
// @Produce json
// @Param id path int true "恢复任务ID"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup/restore/{id} [delete]
func (h *BackupController) DeleteRestoreTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	if err := h.backupService.DeleteRestoreTask(uint(id)); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Restore task deleted successfully"))
}

// DownloadBackup 下载备份文件
// @Summary 下载备份文件
// @Description 下载指定的备份文件
// @Tags backup
// @Accept json
// @Produce application/octet-stream
// @Param id path int true "备份ID"
// @Success 200 {file} binary "备份文件"
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup/download/{id} [get]
func (h *BackupController) DownloadBackup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 获取备份任务信息
	backupTask, err := h.backupService.GetBackupTaskByID(uint(id))
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 检查备份文件路径是否存在
	if backupTask.StoragePath == "" || backupTask.Status != "completed" {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBackupNotFound)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 获取文件信息
	fileInfo, err := os.Stat(backupTask.StoragePath)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBackupNotFound)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileInfo.Name())
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// 发送文件
	c.File(backupTask.StoragePath)
}

// UploadBackup 上传备份文件
// @Summary 上传备份文件
// @Description 上传备份文件压缩包，上传完成后会添加记录到备份表中，上传逻辑异步
// @Tags backup
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "备份文件压缩包"
// @Success 200 {object} success.DataResponse{data=models.BackupTask}
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/admin/backup/upload [post]
func (h *BackupController) UploadBackup(c *gin.Context) {
	startTime := time.Now()
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	// 创建备份任务
	task, err := h.backupService.CreateUploadBackupTask(startTime, file)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Backup upload task created successfully", task))
}
