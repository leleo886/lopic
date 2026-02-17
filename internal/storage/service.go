package storage

import (
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/internal/config"
)

// StorageService 存储服务
type StorageService struct {
	config *config.Config
}

// NewStorageService 创建存储服务实例
func NewStorageService(config *config.Config) *StorageService {
	return &StorageService{
		config: config,
	}
}

// GetStorageByStorage 根据存储配置获取存储实例
func (s *StorageService) GetStorageByStorage(storage *models.Storage) Storage {
	return NewStorageByStorageName(storage, &s.config.Server)
}

// GetStorageByType 根据存储类型获取存储实例
func (s *StorageService) GetStorageByType(storageType string) Storage {
	// 根据存储类型返回对应的存储实现
	return NewLocalStorage(s.config.Server.UploadDir, s.config.Server.StaticPath)
}
