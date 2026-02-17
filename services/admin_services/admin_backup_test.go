package admin_services

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/models"
)

func TestNewBackupService(t *testing.T) {
	tests := []struct {
		name         string
		dbConfig     *config.DatabaseConfig
		serverConfig *config.ServerConfig
		expectNil    bool
	}{
		{
			name:         "with all configs",
			dbConfig:     &config.DatabaseConfig{Type: "sqlite"},
			serverConfig: &config.ServerConfig{UploadDir: "uploads"},
			expectNil:    false,
		},
		{
			name:         "with nil configs",
			dbConfig:     nil,
			serverConfig: nil,
			expectNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewBackupService(nil, nil, tt.dbConfig, tt.serverConfig)
			if (service == nil) != tt.expectNil {
				t.Errorf("NewBackupService() returned nil = %v, want nil = %v", service == nil, tt.expectNil)
			}
		})
	}
}

func TestGetDatabaseDriver(t *testing.T) {
	tests := []struct {
		name         string
		dbConfig     *config.DatabaseConfig
		expectedType string
	}{
		{
			name:         "sqlite driver",
			dbConfig:     &config.DatabaseConfig{Type: "sqlite"},
			expectedType: "sqlite",
		},
		{
			name:         "mysql driver",
			dbConfig:     &config.DatabaseConfig{Type: "mysql"},
			expectedType: "mysql",
		},
		{
			name:         "nil config",
			dbConfig:     nil,
			expectedType: "unknown",
		},
		{
			name:         "empty type",
			dbConfig:     &config.DatabaseConfig{Type: ""},
			expectedType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewBackupService(nil, nil, tt.dbConfig, nil)
			driver := service.getDatabaseDriver()
			if driver != tt.expectedType {
				t.Errorf("getDatabaseDriver() = %v, want %v", driver, tt.expectedType)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcContent := []byte("test content for copy file")
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "destination.txt")

	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Errorf("copyFile() error = %v", err)
		return
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("failed to read destination file: %v", err)
		return
	}

	if !bytes.Equal(dstContent, srcContent) {
		t.Errorf("copyFile() content = %v, want %v", dstContent, srcContent)
	}
}

func TestCopyFileNonExistentSource(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "nonexistent.txt")
	dstPath := filepath.Join(tempDir, "destination.txt")

	err = copyFile(srcPath, dstPath)
	if err == nil {
		t.Error("copyFile() expected error for non-existent source, got nil")
	}
}

func TestExtractBackup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	backupPath := filepath.Join(tempDir, "test_backup.zip")

	zipFile, err := os.Create(backupPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	files := []struct {
		name    string
		content string
	}{
		{"database_sqlite.sql", "CREATE TABLE test (id INTEGER);"},
		{"uploads/image1.jpg", "fake image content"},
		{"uploads/subdir/image2.png", "another fake image"},
	}

	for _, f := range files {
		w, err := zipWriter.Create(f.name)
		if err != nil {
			t.Fatalf("failed to create zip entry: %v", err)
		}
		if _, err := w.Write([]byte(f.content)); err != nil {
			t.Fatalf("failed to write zip content: %v", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("failed to close zip file: %v", err)
	}

	service := NewBackupService(nil, nil, nil, nil)
	extractDir, err := service.extractBackup(backupPath)
	if err != nil {
		t.Errorf("extractBackup() error = %v", err)
		return
	}
	defer os.RemoveAll(extractDir)

	for _, f := range files {
		extractedPath := filepath.Join(extractDir, f.name)
		content, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("failed to read extracted file %s: %v", f.name, err)
			continue
		}
		if string(content) != f.content {
			t.Errorf("extractBackup() file %s content = %v, want %v", f.name, string(content), f.content)
		}
	}
}

func TestExtractBackupInvalidZip(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidZipPath := filepath.Join(tempDir, "invalid.zip")
	if err := os.WriteFile(invalidZipPath, []byte("not a valid zip file"), 0644); err != nil {
		t.Fatalf("failed to create invalid zip file: %v", err)
	}

	service := NewBackupService(nil, nil, nil, nil)
	_, err = service.extractBackup(invalidZipPath)
	if err == nil {
		t.Error("extractBackup() expected error for invalid zip, got nil")
	}
}

func TestExtractBackupNonExistent(t *testing.T) {
	service := NewBackupService(nil, nil, nil, nil)
	_, err := service.extractBackup("/nonexistent/path/backup.zip")
	if err == nil {
		t.Error("extractBackup() expected error for non-existent file, got nil")
	}
}

func TestBackupTaskStatus(t *testing.T) {
	task := models.BackupTask{
		Status:    "pending",
		StartTime: parseTime("2024-01-01T00:00:00Z"),
	}

	if task.Status != "pending" {
		t.Errorf("expected status pending, got %s", task.Status)
	}

	task.Status = "running"
	if task.Status != "running" {
		t.Errorf("expected status running, got %s", task.Status)
	}

	task.Status = "completed"
	if task.Status != "completed" {
		t.Errorf("expected status completed, got %s", task.Status)
	}
}

func TestRestoreTaskStatus(t *testing.T) {
	task := models.RestoreTask{
		BackupTaskID: 1,
		Status:       "pending",
		StartTime:    parseTime("2024-01-01T00:00:00Z"),
	}

	if task.Status != "pending" {
		t.Errorf("expected status pending, got %s", task.Status)
	}

	if task.BackupTaskID != 1 {
		t.Errorf("expected BackupTaskID 1, got %d", task.BackupTaskID)
	}
}

func TestBackupTaskWithEndTime(t *testing.T) {
	endTime := parseTime("2024-01-01T01:00:00Z")
	task := models.BackupTask{
		Status:    "completed",
		StartTime: parseTime("2024-01-01T00:00:00Z"),
		EndTime:   &endTime,
		Size:      1024000,
	}

	if task.EndTime == nil {
		t.Error("expected EndTime to be set")
	}

	if task.Size != 1024000 {
		t.Errorf("expected size 1024000, got %d", task.Size)
	}
}

func TestBackupServiceWithNilDependencies(t *testing.T) {
	service := NewBackupService(nil, nil, nil, nil)

	if service.db != nil {
		t.Error("expected db to be nil")
	}

	if service.storageService != nil {
		t.Error("expected storageService to be nil")
	}

	if service.dbConfig != nil {
		t.Error("expected dbConfig to be nil")
	}

	if service.serverConfig != nil {
		t.Error("expected serverConfig to be nil")
	}
}

func TestBackupDatabaseUnsupportedDriver(t *testing.T) {
	service := NewBackupService(nil, nil, &config.DatabaseConfig{Type: "postgres"}, nil)

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	err := service.backupDatabase(zipWriter)
	if err == nil {
		t.Error("backupDatabase() expected error for unsupported driver, got nil")
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInternalServer",
			err:      cerrors.ErrInternalServer,
			expected: "internal server error",
		},
		{
			name:     "ErrBackupDatabase",
			err:      cerrors.ErrBackupDatabase,
			expected: "error backing up database",
		},
		{
			name:     "ErrBackupFiles",
			err:      cerrors.ErrBackupFiles,
			expected: "error backing up files",
		},
		{
			name:     "ErrExtractBackup",
			err:      cerrors.ErrExtractBackup,
			expected: "error extracting backup file",
		},
		{
			name:     "ErrNoValidDBBackup",
			err:      cerrors.ErrNoValidDBBackup,
			expected: "no valid database backup file found in archive",
		},
		{
			name:     "ErrDBTypeMismatch",
			err:      cerrors.ErrDBTypeMismatch,
			expected: "backup database type does not match current database type",
		},
		{
			name:     "ErrMySQLRestore",
			err:      cerrors.ErrMySQLRestore,
			expected: "error restoring MySQL database",
		},
		{
			name:     "ErrSQLiteRestore",
			err:      cerrors.ErrSQLiteRestore,
			expected: "error restoring SQLite database",
		},
		{
			name:     "ErrRestoreFiles",
			err:      cerrors.ErrRestoreFiles,
			expected: "error restoring files",
		},
		{
			name:     "ErrBackupTaskNotCompleted",
			err:      cerrors.ErrBackupTaskNotCompleted,
			expected: "backup task status is not completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("error message = %v, want %v", tt.err.Error(), tt.expected)
			}
		})
	}
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
