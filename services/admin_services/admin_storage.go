package admin_services

import (
	"fmt"

	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/storage"
	"github.com/leleo886/lopic/models"
	"gorm.io/gorm"
)

type StorageService struct {
	db *gorm.DB
}

func NewStorageService(db *gorm.DB) *StorageService {
	return &StorageService{db: db}
}

type StorageRequest struct {
	Name   string               `json:"name"`
	Type   string               `json:"type"`
	Config models.StorageConfig `json:"config"`
}

func (s *StorageService) GetStorages(page, pageSize, offset int, searchkey, orderby, order string) (*[]models.Storage, error) {
	var storages []models.Storage
	query := s.db
	if searchkey != "" {
		sk := fmt.Sprintf("%%%s%%", searchkey)
		query = query.Where("name LIKE ? OR type LIKE ?", sk, sk)
	}
	result := query.Limit(pageSize).Offset(offset).
		Order(fmt.Sprintf("%s %s", orderby, order)).Find(&storages)
	if result.Error != nil {
		log.Errorf("failed to get storages: error=%v", result.Error)
		return nil, cerrors.ErrInternalServer
	}
	return &storages, nil
}

func (s *StorageService) GetStorage(id int) (*models.Storage, error) {
	var storage models.Storage
	result := s.db.Where("id = ?", id).First(&storage)
	if result.Error != nil {
		return nil, cerrors.ErrStorageNotFound
	}
	return &storage, nil
}

func (s *StorageService) GetStorageByName(name string) (*models.Storage, error) {
	var storage models.Storage
	result := s.db.Where("name = ?", name).First(&storage)
	if result.Error != nil {
		return nil, cerrors.ErrStorageNotFound
	}
	return &storage, nil
}

func (s *StorageService) CreateStorage(storageReq *StorageRequest) error {
	// 检查 name 是否已存在
	var existingStorage models.Storage
	result := s.db.Where("name = ?", storageReq.Name).First(&existingStorage)
	if result.RowsAffected > 0 {
		return cerrors.ErrStorageNameExists
	}

	// 检查是否要创建 local 类型的存储
	if storageReq.Type == "local" {
		// 检查是否已存在 local 类型的存储
		var localStorage models.Storage
		localResult := s.db.Where("type = ?", "local").First(&localStorage)
		if localResult.RowsAffected > 0 {
			return cerrors.ErrBadRequest
		}
	}

	// 创建存储配置
	storage := &models.Storage{
		Name:   storageReq.Name,
		Type:   storageReq.Type,
		Config: storageReq.Config,
	}

	result = s.db.Create(storage)
	if result.Error != nil {
		log.Errorf("failed to create storage: error=%v", result.Error)
		return cerrors.ErrInternalServer
	}
	return nil
}

func (s *StorageService) UpdateStorage(id uint, storageReq *StorageRequest) error {
	var existingStorage models.Storage
	result := s.db.Where("id = ?", id).First(&existingStorage)
	if result.Error != nil {
		return cerrors.ErrStorageNotFound
	}

	// local 存储不可修改名称和类型
	if existingStorage.Name == "local" {
		if storageReq.Name != "local" {
			return cerrors.ErrCannotChangeLocalStorageName
		}
		if storageReq.Type != "local" {
			return cerrors.ErrCannotChangeLocalStorageType
		}
	} else {
		// 检查新名称是否与其他存储冲突
		var otherStorage models.Storage
		result = s.db.Where("name = ? AND id != ?", storageReq.Name, id).First(&otherStorage)
		if result.RowsAffected > 0 {
			return cerrors.ErrStorageNameExists
		}

		// 检查是否要将其他类型的存储修改为 local 类型
		if storageReq.Type == "local" {
			return cerrors.ErrBadRequest
		}
	}

	// 更新存储配置
	existingStorage.Name = storageReq.Name
	existingStorage.Type = storageReq.Type

	// 只在提供了新密码时更新密码
	if storageReq.Config.Password != "" {
		// 更新所有配置，包括密码
		existingStorage.Config = storageReq.Config
	} else {
		// 只更新非密码配置
		existingStorage.Config.BaseURL = storageReq.Config.BaseURL
		existingStorage.Config.Username = storageReq.Config.Username
		existingStorage.Config.StaticURL = storageReq.Config.StaticURL
		existingStorage.Config.BasePath = storageReq.Config.BasePath
	}

	result = s.db.Save(&existingStorage)
	if result.Error != nil {
		log.Errorf("failed to update storage: id=%d, error=%v", id, result.Error)
		return cerrors.ErrInternalServer
	}
	return nil
}

func (s *StorageService) TestStorageConnection(storageReq *StorageRequest) error {
	var existingStorage models.Storage
	result := s.db.Where("name = ?", storageReq.Name).First(&existingStorage)
	if result.Error != nil {
		return cerrors.ErrStorageNotFound
	}

	// 根据存储类型创建存储实例并测试连接
	var storageInstance storage.Storage

	switch storageReq.Type {
	case "webdav":
		storageInstance = storage.NewWebDAVStorage(
			storageReq.Config.BaseURL,
			storageReq.Config.Username,
			existingStorage.Config.Password,
			storageReq.Config.StaticURL,
			storageReq.Config.BasePath,
		)
	case "local":
		storageInstance = storage.NewLocalStorage(
			storageReq.Config.BasePath,
			storageReq.Config.StaticURL,
		)
	default:
		return cerrors.ErrBadRequest
	}

	// 测试连接
	return storageInstance.TestConnection()
}

func (s *StorageService) DeleteStorage(id uint) error {
	var existingStorage models.Storage
	result := s.db.Where("id = ?", id).First(&existingStorage)
	if result.Error != nil {
		return cerrors.ErrStorageNotFound
	}

	// local 存储不可删除
	if existingStorage.Type == "local" {
		return cerrors.ErrCannotDeleteLocalStorage
	}

	// 检查是否有角色引用该存储
	var roles []models.Role
	result = s.db.Where("storage_name = ?", existingStorage.Name).Find(&roles)
	if result.RowsAffected > 0 {
		return cerrors.ErrStorageAssociatedWithRoles
	}

	// 暂不检查图片引用，若重新添加已删除的存储配置可恢复图片删除操作
	var images []models.Image
	result = s.db.Where("storage_name = ?", existingStorage.Name).Find(&images)
	if result.RowsAffected > 0 {
		return cerrors.ErrStorageAssociatedWithImages
	}

	// 删除存储
	result = s.db.Delete(&existingStorage)
	if result.Error != nil {
		log.Errorf("failed to delete storage: id=%d, error=%v", id, result.Error)
		return cerrors.ErrInternalServer
	}
	return nil
}
