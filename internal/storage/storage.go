package storage

import (
	"mime/multipart"
)

// Storage 定义存储接口
type Storage interface {
	// UploadFile 上传文件到存储系统
	UploadFile(file *multipart.FileHeader, filePath string, uploadPath string, fileName string) (string, error)
	// DeleteFile 从存储系统删除文件
	DeleteFile(filePath string) error
	// CreateDirectory 创建目录
	CreateDirectory(dirPath string) error
	// TestConnection 测试存储连接是否成功
	TestConnection() error
}
