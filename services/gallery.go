package services

import (
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"gorm.io/gorm"
)

type GalleryService struct {
	db *gorm.DB
}

func NewGalleryService(db *gorm.DB) *GalleryService {
	return &GalleryService{db: db}
}

type GalleryImageResponse struct {
	FileName        string   `json:"file_name"`
	OriginalName    string   `json:"original_name"`
	Tags            []string `json:"tags" gorm:"serializer:json"`
	FileURL         string   `json:"file_url"`
	FileSize        int64    `json:"file_size"`
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	ThumbnailURL    string   `json:"thumbnail_url"`
	ThumbnailSize   int64    `json:"thumbnail_size"`
	ThumbnailWidth  int      `json:"thumbnail_width"`
	ThumbnailHeight int      `json:"thumbnail_height"`
	MimeType        string   `json:"mime_type"`
}

type GetGalleryImagesResponse struct {
	Images   []GalleryImageResponse `json:"images"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

func (s *GalleryService) GetGallery(currentUserID uint) (*GetAlbumsResponse, error) {
	var albums []AlbumResponse
	var total int64

	result := s.db.Where("user_id = ? AND gallery_enabled = ?", currentUserID, true).
		Order("serial_number ASC").
		Find(&models.Album{})
	result.Count(&total).Find(&albums)
	if result.Error != nil {
		log.Errorf("failed to get albums: user_id=%d, error=%v", currentUserID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	return &GetAlbumsResponse{
		Albums: albums,
		Total:  total,
	}, nil
}

func (s *GalleryService) GetGalleryImages(currentUserID uint, albumID uint, page, pageSize, offset int) (*GetGalleryImagesResponse, error) {
	var album models.Album
	result := s.db.Where("id = ? AND user_id = ?", albumID, currentUserID).First(&album)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrAlbumNotFound
	}

	var AlbumImagesResponse []GalleryImageResponse
	var total int64

	db := s.db.Model(&models.Image{}).
		Joins("JOIN image_albums ON image_albums.image_id = images.id").
		Where("image_albums.album_id = ?", albumID)
	db.Count(&total)
	db.Offset(offset).Limit(pageSize).Order("image_albums.image_id DESC").Find(&AlbumImagesResponse)

	return &GetGalleryImagesResponse{
		Images:   AlbumImagesResponse,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *GalleryService) SearchGalleryImages(currentUserID uint, query string, page, pageSize, offset int) (*GetGalleryImagesResponse, error) {
	var images []GalleryImageResponse
	var total int64

	db := s.db.Model(&models.Image{}).
		Joins("JOIN image_albums ON image_albums.image_id = images.id").
		Where("image_albums.album_id IN (SELECT id FROM albums WHERE user_id = ?)", currentUserID)
	if query != "" {
		db.Where("images.original_name LIKE ? OR images.tags LIKE ?", "%"+query+"%", "%"+query+"%")
	}
	db.Count(&total)
	db.Offset(offset).Limit(pageSize).Order("image_albums.image_id DESC").Find(&images)

	return &GetGalleryImagesResponse{
		Images:   images,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
