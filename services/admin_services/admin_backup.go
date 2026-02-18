package admin_services

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/storage"
	"github.com/leleo886/lopic/models"
	"golang.org/x/text/encoding/simplifiedchinese"
	"gorm.io/gorm"
)

type BackupService struct {
	db             *gorm.DB
	storageService *storage.StorageService
	dbConfig       *config.DatabaseConfig
	serverConfig   *config.ServerConfig
}

func NewBackupService(db *gorm.DB, storageService *storage.StorageService, dbConfig *config.DatabaseConfig, serverConfig *config.ServerConfig) *BackupService {
	return &BackupService{
		db:             db,
		storageService: storageService,
		dbConfig:       dbConfig,
		serverConfig:   serverConfig,
	}
}

func (s *BackupService) getDatabaseDriver() string {
	if s.dbConfig != nil && s.dbConfig.Type != "" {
		return s.dbConfig.Type
	}
	return "unknown"
}

func (s *BackupService) CreateBackup() (*models.BackupTask, error) {
	task := &models.BackupTask{
		Status:    "pending",
		StartTime: time.Now(),
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, cerrors.ErrInternalServer
	}

	// 启动异步备份任务
	go func(taskID uint) {
		if err := s.executeBackup(taskID); err != nil {
			s.updateTaskStatus(taskID, "failed", err.Error())
			log.Errorf("BackupTask %d failed: %v", taskID, err)
		}
	}(task.ID)

	return task, nil
}

func (s *BackupService) executeBackup(taskID uint) error {
	// 更新任务状态为运行中
	var task models.BackupTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	task.Status = "running"
	if err := s.db.Save(&task).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	backupDir := "data/backup"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Errorf("Create backup directory failed: %v", err)
		return cerrors.ErrInternalServer
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFilename := fmt.Sprintf("backup_%s.zip", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	backupFile, err := os.Create(backupPath)
	if err != nil {
		log.Errorf("Create backup file failed: %v", err)
		return cerrors.ErrInternalServer
	}
	defer backupFile.Close()

	zipWriter := zip.NewWriter(backupFile)
	defer zipWriter.Close()

	if err := s.backupDatabase(zipWriter); err != nil {
		log.Errorf("Backup database failed: %v", err)
		return cerrors.ErrBackupDatabase
	}

	if err := s.backupFiles(zipWriter); err != nil {
		log.Errorf("Backup files failed: %v", err)
		return cerrors.ErrBackupFiles
	}

	if err := zipWriter.Close(); err != nil {
		log.Errorf("Close zip writer failed: %v", err)
		return cerrors.ErrInternalServer
	}

	if err := backupFile.Close(); err != nil {
		log.Errorf("Close backup file failed: %v", err)
		return cerrors.ErrInternalServer
	}

	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		log.Errorf("Get backup file info failed: %v", err)
		return cerrors.ErrInternalServer
	}

	endTime := time.Now()
	task.Status = "completed"
	task.EndTime = &endTime
	task.Size = fileInfo.Size()
	task.StoragePath = backupPath

	if err := s.db.Save(&task).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	return nil
}

func (s *BackupService) GetBackupList() ([]models.BackupTask, error) {
	var tasks []models.BackupTask
	if err := s.db.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, cerrors.ErrInternalServer
	}
	return tasks, nil
}

func (s *BackupService) GetRestoreRecords() ([]models.RestoreTask, error) {
	var records []models.RestoreTask
	if err := s.db.Order("created_at DESC").Preload("BackupTask").Find(&records).Error; err != nil {
		return nil, cerrors.ErrInternalServer
	}
	return records, nil
}

// GetBackupTaskByID 根据ID获取备份任务
func (s *BackupService) GetBackupTaskByID(id uint) (models.BackupTask, error) {
	var task models.BackupTask
	if err := s.db.First(&task, id).Error; err != nil {
		return models.BackupTask{}, cerrors.ErrInternalServer
	}
	return task, nil
}

// CreateUploadBackupTask 创建上传备份任务
func (s *BackupService) CreateUploadBackupTask(startTime time.Time, file *multipart.FileHeader) (*models.BackupTask, error) {
	task := &models.BackupTask{
		Status:    "pending",
		StartTime: startTime,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, cerrors.ErrInternalServer
	}

	// 先保存上传的文件到临时目录，避免 HTTP 请求结束后临时文件被删除
	tempDir := "data/temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		s.db.Delete(task)
		log.Errorf("Create temp directory failed: %v", err)
		return nil, cerrors.ErrInternalServer
	}

	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("upload_%d_%s", task.ID, file.Filename))
	src, err := file.Open()
	if err != nil {
		s.db.Delete(task)
		log.Errorf("Open uploaded file failed: %v", err)
		return nil, cerrors.ErrInternalServer
	}
	defer src.Close()

	dst, err := os.Create(tempFilePath)
	if err != nil {
		s.db.Delete(task)
		src.Close()
		log.Errorf("Create temp file failed: %v", err)
		return nil, cerrors.ErrInternalServer
	}

	if _, err = io.Copy(dst, src); err != nil {
		s.db.Delete(task)
		dst.Close()
		os.Remove(tempFilePath)
		log.Errorf("Copy uploaded file to temp failed: %v", err)
		return nil, cerrors.ErrInternalServer
	}
	dst.Close()

	// 启动异步上传任务
	go func(taskID uint, tempFilePath string) {
		defer os.Remove(tempFilePath) // 处理完成后删除临时文件
		if err := s.executeUploadBackup(taskID, tempFilePath); err != nil {
			s.updateTaskStatus(taskID, "failed", err.Error())
			log.Errorf("UploadBackupTask %d failed: %v", taskID, err)
		}
	}(task.ID, tempFilePath)

	return task, nil
}

// executeUploadBackup 执行上传备份任务
func (s *BackupService) executeUploadBackup(taskID uint, tempFilePath string) error {
	// 更新任务状态为运行中
	var task models.BackupTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	task.Status = "running"
	if err := s.db.Save(&task).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	backupDir := "data/backup"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Errorf("Create backup directory failed: %v", err)
		return cerrors.ErrInternalServer
	}

	// 生成备份文件路径
	timestamp := time.Now().Format("20060102_150405")
	backupFilename := fmt.Sprintf("backup_%s.zip", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	// 移动临时文件到备份目录
	if err := os.Rename(tempFilePath, backupPath); err != nil {
		// 如果跨设备移动失败，使用复制+删除
		src, err := os.Open(tempFilePath)
		if err != nil {
			log.Errorf("Open temp file failed: %v", err)
			return cerrors.ErrInternalServer
		}
		defer src.Close()

		dst, err := os.Create(backupPath)
		if err != nil {
			log.Errorf("Create backup file failed: %v", err)
			return cerrors.ErrInternalServer
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			log.Errorf("Copy backup file failed: %v", err)
			return cerrors.ErrInternalServer
		}
	}

	// 获取文件大小
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		log.Errorf("Get backup file info failed: %v", err)
		return cerrors.ErrInternalServer
	}

	// 更新任务状态为完成
	endTime := time.Now()
	task.Status = "completed"
	task.EndTime = &endTime
	task.Size = fileInfo.Size()
	task.StoragePath = backupPath

	if err := s.db.Save(&task).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	return nil
}

func (s *BackupService) DeleteRestoreTask(restoreID uint) error {
	var task models.RestoreTask
	if err := s.db.First(&task, restoreID).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	if err := s.db.Delete(&task).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	return nil
}

func (s *BackupService) DeleteBackup(backupID uint) error {
	var task models.BackupTask
	if err := s.db.First(&task, backupID).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	// 先删除引用该备份的所有恢复任务
	if err := s.db.Where("backup_task_id = ?", backupID).Delete(&models.RestoreTask{}).Error; err != nil {
		log.Errorf("Delete restore tasks failed: %v", err)
		return cerrors.ErrInternalServer
	}

	// 删除备份文件
	if task.StoragePath != "" {
		if err := os.Remove(task.StoragePath); err != nil && !os.IsNotExist(err) {
			log.Errorf("Delete backup file failed: %v", err)
			return cerrors.ErrInternalServer
		}
	}

	// 删除备份任务
	if err := s.db.Delete(&task).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	return nil
}

func (s *BackupService) RestoreBackup(backupID uint) (*models.RestoreTask, error) {
	restoreTask := &models.RestoreTask{
		BackupTaskID: backupID,
		Status:       "pending",
		StartTime:    time.Now(),
	}

	if err := s.db.Create(restoreTask).Error; err != nil {
		return nil, cerrors.ErrInternalServer
	}

	// 启动异步恢复任务
	go func(taskID uint) {
		if err := s.executeRestore(taskID); err != nil {
			log.Errorf("RestoreTask %d failed: %v", taskID, err)
			s.updateRestoreTaskStatus(taskID, "failed", err.Error())
		}
	}(restoreTask.ID)

	return restoreTask, nil
}

func (s *BackupService) executeRestore(taskID uint) error {
	// 更新任务状态为运行中
	var restoreTask models.RestoreTask
	if err := s.db.First(&restoreTask, taskID).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	restoreTask.Status = "running"
	if err := s.db.Save(&restoreTask).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	// 只读方式加载备份记录，避免意外修改
	var backupTask models.BackupTask
	if err := s.db.Raw("SELECT * FROM backup_tasks WHERE id = ?", restoreTask.BackupTaskID).Scan(&backupTask).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	if backupTask.Status != "completed" {
		log.Errorf("Backup task status is not completed: %s", backupTask.Status)
		return cerrors.ErrBackupTaskNotCompleted
	}

	if _, err := os.Stat(backupTask.StoragePath); os.IsNotExist(err) {
		log.Errorf("Backup file does not exist: %s", backupTask.StoragePath)
		return cerrors.ErrBackupFiles
	}

	extractDir, err := s.extractBackup(backupTask.StoragePath)
	if err != nil {
		log.Errorf("Extract backup file failed: %v", err)
		return cerrors.ErrExtractBackup
	}
	defer os.RemoveAll(extractDir)

	mysqlFile := filepath.Join(extractDir, "database_mysql.sql")
	sqliteFile := filepath.Join(extractDir, "database_sqlite.sql")

	var dbType string
	if _, err := os.Stat(mysqlFile); err == nil {
		dbType = "mysql"
	} else if _, err := os.Stat(sqliteFile); err == nil {
		dbType = "sqlite"
	} else {
		log.Errorf("No valid database backup file found in archive")
		return cerrors.ErrNoValidDBBackup
	}

	currentDriver := s.getDatabaseDriver()
	if currentDriver != dbType {
		log.Errorf("Database type mismatch: backup is %s but current is %s", dbType, currentDriver)
		return cerrors.ErrDBTypeMismatch
	}

	switch dbType {
	case "mysql":
		if err := s.restoreMySQL(extractDir); err != nil {
			log.Errorf("MySQL restore failed: %v", err)
			return cerrors.ErrMySQLRestore
		}
	case "sqlite":
		if err := s.restoreSQLite(extractDir); err != nil {
			log.Errorf("SQLite restore failed: %v", err)
			return cerrors.ErrSQLiteRestore
		}
	}

	if err := s.restoreFiles(extractDir); err != nil {
		log.Errorf("Restore files failed: %v", err)
		return cerrors.ErrRestoreFiles
	}

	endTime := time.Now()
	restoreTask.Status = "completed"
	restoreTask.EndTime = &endTime

	if err := s.db.Save(&restoreTask).Error; err != nil {
		return cerrors.ErrInternalServer
	}

	return nil
}

// decodeFilename 解码 ZIP 文件名，支持 UTF-8 和 GBK 编码
func decodeFilename(name string, nonUTF8 bool) string {
	if nonUTF8 {
		decoded, err := simplifiedchinese.GBK.NewDecoder().String(name)
		if err == nil {
			return decoded
		}
	}
	return name
}

func (s *BackupService) extractBackup(backupPath string) (string, error) {
	extractDir, err := os.MkdirTemp("", "restore_*")
	if err != nil {
		log.Errorf("Create temp dir failed: %v", err)
		return "", cerrors.ErrInternalServer
	}

	// 获取解压目录的绝对路径，用于后续的安全检查
	extractDirAbs, err := filepath.Abs(extractDir)
	if err != nil {
		log.Errorf("Get absolute path failed: %v", err)
		return "", cerrors.ErrInternalServer
	}

	zipFile, err := zip.OpenReader(backupPath)
	if err != nil {
		log.Errorf("Open backup file failed: %v", err)
		return "", cerrors.ErrInternalServer
	}
	defer zipFile.Close()

	for _, file := range zipFile.File {
		// 解码文件名，处理不同编码
		fileName := decodeFilename(file.Name, file.NonUTF8)

		// 安全检查：防止 Zip Slip 攻击
		// 清理文件名并检查路径遍历
		cleanName := filepath.Clean(fileName)

		// 拒绝绝对路径和包含 .. 的路径
		if filepath.IsAbs(cleanName) || strings.Contains(cleanName, "..") {
			log.Errorf("Zip Slip attack detected: invalid file name %s", fileName)
			return "", cerrors.ErrForbidden
		}

		// 构建目标路径
		path := filepath.Join(extractDirAbs, cleanName)

		// 安全检查：确保目标路径在解压目录内
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			log.Errorf("Get absolute path failed: %v", err)
			return "", cerrors.ErrInternalServer
		}

		// 验证路径前缀，防止路径遍历
		if !strings.HasPrefix(pathAbs, extractDirAbs+string(filepath.Separator)) {
			log.Errorf("Zip Slip attack detected: %s is outside of %s", pathAbs, extractDirAbs)
			return "", cerrors.ErrForbidden
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(pathAbs, 0755); err != nil {
				log.Errorf("Create dir failed: %v", err)
				return "", cerrors.ErrInternalServer
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(pathAbs), 0755); err != nil {
			log.Errorf("Create dir failed: %v", err)
			return "", cerrors.ErrInternalServer
		}

		dstFile, err := os.OpenFile(pathAbs, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Errorf("Open file failed: %v", err)
			return "", cerrors.ErrInternalServer
		}
		defer dstFile.Close()

		srcFile, err := file.Open()
		if err != nil {
			log.Errorf("Open file failed: %v", err)
			return "", cerrors.ErrInternalServer
		}
		defer srcFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			log.Errorf("Copy file failed: %v", err)
			return "", cerrors.ErrInternalServer
		}
	}

	return extractDir, nil
}

func (s *BackupService) restoreMySQL(extractDir string) error {
	sqlFile := filepath.Join(extractDir, "database_mysql.sql")
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		log.Errorf("SQL file does not exist: %s", sqlFile)
		return fmt.Errorf("SQL file does not exist: %s", sqlFile)
	}

	if s.dbConfig == nil {
		log.Errorf("Database config is nil")
		return fmt.Errorf("database config is nil")
	}

	cmd := exec.Command(
		"mysql",
		fmt.Sprintf("--user=%s", s.dbConfig.User),
		fmt.Sprintf("--password=%s", s.dbConfig.Password),
		fmt.Sprintf("--host=%s", s.dbConfig.Host),
		fmt.Sprintf("--port=%d", s.dbConfig.Port),
	)

	sqlContent, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Errorf("Read SQL file failed: %v", err)
		return cerrors.ErrInternalServer
	}

	cmd.Stdin = bytes.NewReader(sqlContent)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			log.Errorf("MySQL restore failed: %s", stderr.String())
			return fmt.Errorf("mysql restore failed: %s", stderr.String())
		}
		log.Errorf("MySQL restore failed: %v", err)
		return fmt.Errorf("mysql restore failed: %v", err)
	}

	return nil
}

func (s *BackupService) restoreSQLite(extractDir string) error {
	sqlFile := filepath.Join(extractDir, "database_sqlite.sql")
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		log.Errorf("SQL file does not exist: %s", sqlFile)
		return fmt.Errorf("SQL file does not exist: %s", sqlFile)
	}

	sqlContent, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Errorf("Read SQL file failed: %v", err)
		return cerrors.ErrInternalServer
	}

	tables := []string{
		"users",
		"roles",
		"albums",
		"images",
		"system_settings",
		"refresh_token_blacklist",
		"image_albums",
		"storages",
		"password_reset_codes",
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		tx.Rollback()
		log.Errorf("Disable foreign keys failed: %v", err)
		return fmt.Errorf("disable foreign keys failed: %v", err)
	}

	for _, table := range tables {
		if err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			tx.Rollback()
			log.Errorf("Drop table %s failed: %v", table, err)
			return fmt.Errorf("drop table %s failed: %v", table, err)
		}
	}

	statements := strings.Split(string(sqlContent), ";")
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}
		if err := tx.Exec(statement).Error; err != nil {
			tx.Rollback()
			log.Errorf("Execute SQL failed: %v", err)
			return fmt.Errorf("execute SQL failed: %v", err)
		}
	}

	if err := tx.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		tx.Rollback()
		log.Errorf("Enable foreign keys failed: %v", err)
		return fmt.Errorf("enable foreign keys failed: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("Commit transaction failed: %v", err)
		return fmt.Errorf("commit transaction failed: %v", err)
	}

	return nil
}

func (s *BackupService) restoreFiles(extractDir string) error {
	uploadsDir := filepath.Join(extractDir, "uploads")
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		log.Errorf("Uploads dir does not exist: %s", uploadsDir)
		return nil
	}

	targetUploadsDir := strings.Trim(s.serverConfig.UploadDir, "/")
	if _, err := os.Stat(targetUploadsDir); err == nil {
		if err := os.RemoveAll(targetUploadsDir); err != nil {
			log.Errorf("Remove existing uploads dir failed: %v", err)
			return fmt.Errorf("remove existing uploads dir failed: %v", err)
		}
	}

	if err := os.MkdirAll(targetUploadsDir, 0755); err != nil {
		log.Errorf("Create uploads dir failed: %v", err)
		return fmt.Errorf("create uploads dir failed: %v", err)
	}

	return filepath.Walk(uploadsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Walk uploads dir failed: %v", err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(uploadsDir, path)
		if err != nil {
			log.Errorf("Get relative path failed: %v", err)
			return err
		}

		targetPath := filepath.Join(targetUploadsDir, relPath)

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			log.Errorf("Create dir failed: %v", err)
			return err
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		log.Errorf("Open source file failed: %v", err)
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Errorf("Open destination file failed: %v", err)
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *BackupService) backupDatabase(zipWriter *zip.Writer) error {
	driver := s.getDatabaseDriver()

	switch driver {
	case "mysql":
		return s.backupMySQL(zipWriter)
	case "sqlite":
		return s.backupSQLite(zipWriter)
	default:
		return fmt.Errorf("unsupported database driver: %s, only sqlite and mysql are supported", driver)
	}
}

func (s *BackupService) backupMySQL(zipWriter *zip.Writer) error {
	if s.dbConfig == nil {
		log.Errorf("Database configuration is required for MySQL backup")
		return fmt.Errorf("database configuration is required for MySQL backup")
	}

	cmd := exec.Command(
		"mysqldump",
		fmt.Sprintf("--user=%s", s.dbConfig.User),
		fmt.Sprintf("--password=%s", s.dbConfig.Password),
		fmt.Sprintf("--host=%s", s.dbConfig.Host),
		fmt.Sprintf("--port=%d", s.dbConfig.Port),
		"--databases",
		"--single-transaction",
		"--routines",
		"--triggers",
		"--events",
		"--master-data=2",
		"--flush-logs",
		fmt.Sprintf("--ignore-table=%s.backup_tasks", s.dbConfig.DBName),
		fmt.Sprintf("--ignore-table=%s.restore_tasks", s.dbConfig.DBName),
		s.dbConfig.DBName,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			log.Errorf("mysqldump failed: %s", stderr.String())
			return fmt.Errorf("mysqldump failed: %s", stderr.String())
		}
		log.Errorf("mysqldump failed: %v", err)
		return fmt.Errorf("mysqldump failed: %v", err)
	}

	sqlFile, err := zipWriter.Create("database_mysql.sql")
	if err != nil {
		log.Errorf("Create SQL file failed: %v", err)
		return err
	}

	if _, err := sqlFile.Write(stdout.Bytes()); err != nil {
		return err
	}

	return nil
}

func (s *BackupService) backupSQLite(zipWriter *zip.Writer) error {
	if s.dbConfig == nil || s.dbConfig.GetDSN() == "" {
		log.Errorf("Database configuration is required for SQLite backup")
		return fmt.Errorf("database configuration is required for SQLite backup")
	}

	dbPath := s.dbConfig.GetDSN()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Errorf("SQLite database file not found: %s", dbPath)
		return fmt.Errorf("SQLite database file not found: %s", dbPath)
	}

	sqlFile, err := zipWriter.Create("database_sqlite.sql")
	if err != nil {
		return err
	}

	tables := []string{
		"users",
		"roles",
		"albums",
		"images",
		"system_settings",
		"refresh_token_blacklist",
		"image_albums",
		"storages",
		"password_reset_codes",
	}

	for _, table := range tables {
		var tableName, createTableSQL string
		if err := s.db.Raw(fmt.Sprintf("SELECT name, sql FROM sqlite_master WHERE type='table' AND name='%s'", table)).Row().Scan(&tableName, &createTableSQL); err != nil {
			log.Errorf("Failed to get table schema for %s: %v", table, err)
			continue
		}

		if _, err := sqlFile.Write([]byte(createTableSQL + ";\n\n")); err != nil {
			log.Errorf("Write create table SQL failed for %s: %v", table, err)
			return err
		}

		rows, err := s.db.Raw(fmt.Sprintf("SELECT * FROM %s", table)).Rows()
		if err != nil {
			log.Errorf("Failed to query table %s: %v", table, err)
			continue
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			log.Errorf("Failed to get columns for table %s: %v", table, err)
			continue
		}

		values := make([]sql.RawBytes, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}

			var vals []string
			for _, v := range values {
				if v == nil {
					vals = append(vals, "NULL")
				} else {
					escapedValue := strings.ReplaceAll(string(v), "'", "''")
					vals = append(vals, fmt.Sprintf("'%s'", escapedValue))
				}
			}

			if _, err := sqlFile.Write([]byte(fmt.Sprintf("INSERT INTO %s VALUES (%s);\n", table, strings.Join(vals, ", ")))); err != nil {
				return err
			}
		}

		rows.Close()
	}

	return nil
}

func (s *BackupService) backupFiles(zipWriter *zip.Writer) error {
	uploadDir := s.serverConfig.UploadDir
	return filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Walk local storage failed: %v", err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		zipPath, err := filepath.Rel(uploadDir, path)
		if err != nil {
			log.Errorf("Get relative path failed: %v", err)
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Errorf("Create zip file header failed: %v", err)
			return err
		}

		header.Name = filepath.Join("uploads", zipPath)
		header.Method = zip.Deflate
		header.Flags |= 0x800

		zipFile, err := zipWriter.CreateHeader(header)
		if err != nil {
			log.Errorf("Create zip file failed: %v", err)
			return err
		}

		localFile, err := os.Open(path)
		if err != nil {
			log.Errorf("Open local file failed: %v", err)
			return err
		}
		defer localFile.Close()

		if _, err := io.Copy(zipFile, localFile); err != nil {
			log.Errorf("Copy local file to zip file failed: %v", err)
			return err
		}

		return nil
	})
}

func (s *BackupService) updateTaskStatus(taskID uint, status string, errorMsg string) {
	task := models.BackupTask{}
	if err := s.db.First(&task, taskID).Error; err != nil {
		return
	}

	task.Status = status
	if errorMsg != "" {
		task.Error = errorMsg
	}
	endTime := time.Now()
	task.EndTime = &endTime

	s.db.Save(&task)
}

func (s *BackupService) updateRestoreTaskStatus(taskID uint, status string, errorMsg string) {
	task := models.RestoreTask{}
	if err := s.db.First(&task, taskID).Error; err != nil {
		return
	}

	task.Status = status
	if errorMsg != "" {
		task.Error = errorMsg
	}
	if status == "completed" || status == "failed" {
		endTime := time.Now()
		task.EndTime = &endTime
	}

	s.db.Save(&task)
}
