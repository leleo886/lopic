package admin_services

import (
	"fmt"

	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/services"
	"gorm.io/gorm"
)

type AlbumService struct {
	db *gorm.DB
}

func NewAlbumService(db *gorm.DB) *AlbumService {
	return &AlbumService{db: db}
}

func (s *AlbumService) GetAllAlbums(page, pageSize, offset int, searchkey, orderby, order string) (*services.GetAlbumsResponse, error) {
	var albums []services.AlbumResponse
	var total int64

	query := s.db.Model(&models.Album{})
	if searchkey != "" {
		sk := fmt.Sprintf("%%%s%%", searchkey)
		query = query.Where("name LIKE ? OR description LIKE ?", sk, sk)
	}
	query.Count(&total)
	res := query.Offset(offset).Limit(pageSize).
		Order(fmt.Sprintf("%s %s", orderby, order)).Find(&albums)
	if res.Error != nil {
		log.Errorf("failed to get albums: error=%v", res.Error)
		return nil, cerrors.ErrInternalServer
	}

	return &services.GetAlbumsResponse{
		Albums:   albums,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *AlbumService) GetAlbum(id uint) (*services.AlbumResponse, error) {
	var album services.AlbumResponse
	res := s.db.Model(&models.Album{}).Where("id = ?", id).First(&album)
	if res.Error != nil {
		return nil, cerrors.ErrAlbumNotFound
	}
	return &album, nil
}

func (s *AlbumService) DeleteAlbum(id uint) error {
	// 使用事务处理删除操作
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var album models.Album
	res := tx.Where("id = ?", id).First(&album)
	if res.Error != nil {
		tx.Rollback()
		return cerrors.ErrAlbumNotFound
	}

	tx.Model(&album).Association("Images").Clear()

	res = tx.Delete(&album)
	if res.Error != nil {
		tx.Rollback()
		log.Errorf("failed to delete album: id=%d, error=%v", id, res.Error)
		return cerrors.ErrInternalServer
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		return cerrors.ErrInternalServer
	}

	return nil
}
