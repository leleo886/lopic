package admin_services

import (
	"fmt"

	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/storage"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/services"
	"gorm.io/gorm"
)

type ImageService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewImageService(db *gorm.DB, cfg *config.Config) *ImageService {
	return &ImageService{db: db, cfg: cfg}
}

func (s *ImageService) GetAllImages(page, pageSize, offset int, searchkey, field, value, orderby, order string) (*services.GetImagesResponse, error) {
	var imageModels []models.Image
	var total int64

	db := s.db.Model(&imageModels)
	if searchkey != "" {
		sk := fmt.Sprintf("%%%s%%", searchkey)
		db = db.Where("original_name LIKE ? OR tags LIKE ?", sk, sk)
	}
	// 按字段值过滤
	if field != "" && value != "" {
		// 检查字段是否存在，避免 SQL 注入
		allowedFields := map[string]bool{
			"storage_name": true,
			"album_id":     true,
			"user_id":      true,
		}
		if allowedFields[field] {
			if field == "album_id" {
				db = db.Joins("JOIN image_albums ON images.id = image_albums.image_id").Where("image_albums.album_id = ?", value)
			} else {
				db = db.Where(fmt.Sprintf("%s = ?", field), value)
			}
		}
	}
	db.Count(&total)
	res := db.Preload("Albums").Offset(offset).Limit(pageSize).
		Order(fmt.Sprintf("%s %s", orderby, order)).Find(&imageModels)
	if res.Error != nil {
		log.Errorf("failed to get images: error=%v", res.Error)
		return nil, cerrors.ErrInternalServer
	}

	images := services.MakeImagesWithAlbum(imageModels)

	return &services.GetImagesResponse{
		Images:   images,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ImageService) GetImage(id uint) (*services.ImageResponse, error) {
	var imageModel models.Image
	if err := s.db.Preload("Albums").First(&imageModel, id).Error; err != nil {
		return nil, cerrors.ErrImageNotFound
	}
	imageResponse := services.MakeImageWithAlbum(imageModel)
	return &imageResponse, nil
}

func (s *ImageService) DeleteImage(id uint) error {
	// 使用事务处理删除操作
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var imageModel models.Image
	if err := tx.First(&imageModel, id).Error; err != nil {
		tx.Rollback()
		return cerrors.ErrImageNotFound
	}

	// Get all albums that the image belongs to
	var albums []models.Album
	tx.Model(&imageModel).Association("Albums").Find(&albums)

	// Remove the image from each album
	tx.Model(&imageModel).Association("Albums").Clear()

	// Update image_count for each album
	for _, album := range albums {
		if err := tx.Model(&album).Update("image_count", album.ImageCount-1).Error; err != nil {
			tx.Rollback()
			log.Errorf("failed to update album image count: error=%v", err)
			return cerrors.ErrInternalServer
		}
	}

	// 更新用户的存储空间使用情况
	result := tx.Model(&models.User{}).Where("id = ?", imageModel.UserID).
		Updates(map[string]interface{}{
			"total_size":  gorm.Expr("total_size - ?", imageModel.FileSize+imageModel.ThumbnailSize),
			"image_count": gorm.Expr("image_count - ?", 1),
		})
	if result.Error != nil {
		tx.Rollback()
		log.Errorf("failed to update user storage usage: id=%d, error=%v", imageModel.UserID, result.Error)
		return cerrors.ErrInternalServer
	}

	// 删除数据库中的图片记录
	res := tx.Delete(&imageModel)
	if res.Error != nil {
		tx.Rollback()
		log.Errorf("failed to delete image: id=%d, error=%v", id, res.Error)
		return cerrors.ErrInternalServer
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		return cerrors.ErrInternalServer
	}

	// 获取用户对应的存储实例
	storageInstance, err := s.getStorageByStorageName(imageModel.StorageName)
	if err != nil {
		return err
	}

	// Delete the image from storage
	if err := storageInstance.DeleteFile(imageModel.FileURL); err != nil {
		log.Errorf("failed to delete image file: id=%d, error=%v", id, err)
		return cerrors.ErrInternalServer
	}

	if err := storageInstance.DeleteFile(imageModel.ThumbnailURL); err != nil {
		log.Errorf("failed to delete image thumbnail: id=%d, error=%v", id, err)
		return cerrors.ErrInternalServer
	}

	return nil
}

// 根据存储名称获取存储实例
func (s *ImageService) getStorageByStorageName(storageName string) (storage.Storage, error) {
	var storageConfig models.Storage
	result := s.db.Where("name = ?", storageName).First(&storageConfig)
	if result.Error != nil {
		// 存储配置不存在，使用默认本地存储
		return storage.NewStorageByStorageName(nil, &s.cfg.Server), nil
	}

	return storage.NewStorageByStorageName(&storageConfig, &s.cfg.Server), nil
}

// UpdateImageStorage 更新图片存储名称
func (s *ImageService) UpdateImageStorage(id uint, storageName string) error {
	var imageModel models.Image
	if err := s.db.First(&imageModel, id).Error; err != nil {
		return cerrors.ErrImageNotFound
	}

	// 检查存储名称是否存在
	var storageConfig models.Storage
	if err := s.db.Where("name = ?", storageName).First(&storageConfig).Error; err != nil {
		return cerrors.ErrStorageNotFound
	}

	// 更新存储名称
	imageModel.StorageName = storageName
	if err := s.db.Save(&imageModel).Error; err != nil {
		log.Errorf("failed to update image storage: id=%d, error=%v", id, err)
		return cerrors.ErrInternalServer
	}

	return nil
}
