package services

import (
	"time"

	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"gorm.io/gorm"
)

type AlbumService struct {
	db *gorm.DB
}

func NewAlbumService(db *gorm.DB) *AlbumService {
	return &AlbumService{db: db}
}

type AlbumResponse struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"user_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CoverImage     string    `json:"cover_image"`
	ImageCount     int       `json:"image_count"`
	GalleryEnabled bool      `json:"gallery_enabled"`
	SerialNumber   int       `json:"serial_number"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type GetAlbumsResponse struct {
	Albums   []AlbumResponse `json:"albums"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type GetAlbumImagesResponse struct {
	Images   []ImageResponse `json:"images"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func (s *AlbumService) CreateAlbum(name string, description string, galleryEnabled bool, userID uint, serialNumber int) (*AlbumResponse, error) {
	// 获取用户角色
	var user models.User
	result := s.db.Preload("Role").First(&user, userID)
	if result.Error != nil {
		log.Errorf("failed to get user: id=%d, error=%v", userID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	if !user.Role.GalleryOpen && galleryEnabled {
		return nil, cerrors.ErrGalleryPermissionDenied
	}

	// 先获取用户的相册数量
	var count int64
	result = s.db.Model(&models.Album{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		log.Errorf("failed to count albums: user_id=%d, error=%v", userID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	// -1 means no limit
	if user.Role.MaxAlbumsPerUser != -1 {
		if count >= int64(user.Role.MaxAlbumsPerUser) {
			return nil, cerrors.ErrMaxAlbumsPerUser
		}
	}

	album := models.Album{
		Name:           name,
		Description:    description,
		UserID:         userID,
		GalleryEnabled: galleryEnabled,
		SerialNumber:   serialNumber,
		ImageCount:     0,
	}

	result = s.db.Create(&album)
	if result.Error != nil {
		log.Errorf("failed to create album: user_id=%d, error=%v", userID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	return &AlbumResponse{
		ID:             album.ID,
		UserID:         album.UserID,
		Name:           album.Name,
		Description:    album.Description,
		ImageCount:     album.ImageCount,
		GalleryEnabled: album.GalleryEnabled,
		SerialNumber:   album.SerialNumber,
		CreatedAt:      album.CreatedAt,
		UpdatedAt:      album.UpdatedAt,
	}, nil
}

func (s *AlbumService) GetAlbums(userID uint, page, pageSize, offset int) (*GetAlbumsResponse, error) {
	var albums []AlbumResponse
	var total int64

	result := s.db.Where("user_id = ?", userID).Find(&models.Album{})
	result.Count(&total)
	result.Offset(offset).Order("serial_number ASC").Find(&albums)
	if result.Error != nil {
		log.Errorf("failed to get albums: user_id=%d, error=%v", userID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	return &GetAlbumsResponse{
		Albums:   albums,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AlbumService) GetAlbum(id uint, userID uint) (*AlbumResponse, error) {
	var album models.Album
	result := s.db.Where("id = ? AND user_id = ?", id, userID).First(&album)
	if result.Error != nil {
		return nil, cerrors.ErrAlbumNotFound
	}

	return &AlbumResponse{
		ID:             album.ID,
		UserID:         album.UserID,
		Name:           album.Name,
		Description:    album.Description,
		CoverImage:     album.CoverImage,
		ImageCount:     album.ImageCount,
		GalleryEnabled: album.GalleryEnabled,
		SerialNumber:   album.SerialNumber,
		CreatedAt:      album.CreatedAt,
		UpdatedAt:      album.UpdatedAt,
	}, nil
}

func (s *AlbumService) UpdateAlbum(id uint, userID uint, name string, description string, galleryEnabled bool, serialNumber int) (*AlbumResponse, error) {
	// 获取用户角色
	var user models.User
	result := s.db.Preload("Role").First(&user, userID)
	if result.Error != nil {
		log.Errorf("failed to get user: id=%d, error=%v", userID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	if user.Role.GalleryOpen == false && galleryEnabled == true {
		return nil, cerrors.ErrGalleryPermissionDenied
	}

	var album models.Album
	result = s.db.Where("id = ? AND user_id = ?", id, userID).First(&album)
	if result.Error != nil {
		return nil, cerrors.ErrAlbumNotFound
	}

	album.Name = name
	album.Description = description
	album.GalleryEnabled = galleryEnabled
	album.SerialNumber = serialNumber

	result = s.db.Save(&album)
	if result.Error != nil {
		log.Errorf("failed to update album: id=%d, user_id=%d, error=%v", id, userID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	return &AlbumResponse{
		ID:             album.ID,
		UserID:         album.UserID,
		Name:           album.Name,
		Description:    album.Description,
		CoverImage:     album.CoverImage,
		ImageCount:     album.ImageCount,
		GalleryEnabled: album.GalleryEnabled,
		SerialNumber:   album.SerialNumber,
		CreatedAt:      album.CreatedAt,
		UpdatedAt:      album.UpdatedAt,
	}, nil
}

func (s *AlbumService) DeleteAlbum(id uint, userID uint) error {
	// 使用事务处理删除操作
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var album models.Album
	result := tx.Where("id = ? AND user_id = ?", id, userID).First(&album)
	if result.Error != nil {
		tx.Rollback()
		return cerrors.ErrAlbumNotFound
	}

	tx.Model(&album).Association("Images").Clear()

	result = tx.Delete(&album)
	if result.Error != nil {
		tx.Rollback()
		log.Errorf("failed to delete album: id=%d, user_id=%d, error=%v", id, userID, result.Error)
		return cerrors.ErrInternalServer
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		return cerrors.ErrInternalServer
	}

	return nil
}

func (s *AlbumService) GetAlbumImages(id uint, userID uint, page, pageSize, offset int) (*GetAlbumImagesResponse, error) {
	var album models.Album
	result := s.db.Where("id = ? AND user_id = ?", id, userID).First(&album)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrAlbumNotFound
	}

	var imageModels []models.Image
	var AlbumImagesResponse []ImageResponse
	var total int64

	db := s.db.Model(&models.Image{}).
		Joins("JOIN image_albums ON image_albums.image_id = images.id").
		Where("image_albums.album_id = ?", id)
	db.Count(&total)
	db.Offset(offset).Limit(pageSize).Order("image_albums.image_id DESC").Find(&imageModels)

	AlbumImagesResponse = MakeImagesWithAlbum(imageModels)

	return &GetAlbumImagesResponse{
		Images:   AlbumImagesResponse,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AlbumService) GetNotInAnyAlbum(userID uint, page, pageSize, offset int) (*GetAlbumImagesResponse, error) {
	var imageModels []models.Image
	var AlbumImagesResponse []ImageResponse
	var total int64

	db := s.db.Model(&models.Image{}).
		Joins("LEFT JOIN image_albums ON image_albums.image_id = images.id").
		Where("image_albums.album_id IS NULL AND images.user_id = ?", userID)
	db.Count(&total)
	db.Offset(offset).Limit(pageSize).Order("images.id DESC").Find(&imageModels)
	
	AlbumImagesResponse = MakeImagesWithAlbum(imageModels)

	return &GetAlbumImagesResponse{
		Images:   AlbumImagesResponse,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
