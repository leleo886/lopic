package storage

import (
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/models"
)

// NewStorageByStorageName 根据存储名称创建存储实例
func NewStorageByStorageName(storage *models.Storage, config *config.ServerConfig) Storage {
	if storage == nil {
		// 默认使用本地存储
		return NewLocalStorage(
			config.UploadDir,
			config.StaticPath,
		)
	}

	// 存储配置已经是 StorageConfig 类型，直接使用
	storageConfig := storage.Config

	switch storage.Type {
	case "webdav":
		return NewWebDAVStorage(
			storageConfig.BaseURL,
			storageConfig.Username,
			storageConfig.Password,
			storageConfig.StaticURL,
			storageConfig.BasePath,
		)
	case "local":
		fallthrough
	default:
		// 本地存储使用配置文件中的目录
		basePath := config.UploadDir
		staticPath := config.StaticPath

		// 如果存储配置中有自定义路径，使用自定义路径
		// if storageConfig.BasePath != "" {
		// 	basePath = storageConfig.BasePath
		// }
		// if storageConfig.StaticURL != "" {
		// 	staticPath = storageConfig.StaticURL
		// }

		return NewLocalStorage(
			basePath,
			staticPath,
		)
	}
}

