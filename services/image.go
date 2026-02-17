package services

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/storage"
	"github.com/leleo886/lopic/models"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
	"gorm.io/gorm"
)

type ImageService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewImageService(db *gorm.DB, cfg *config.Config) *ImageService {
	return &ImageService{
		db:  db,
		cfg: cfg,
	}
}

type ImageResponse struct {
	ID              uint            `json:"id"`
	FileName        string          `json:"file_name"`
	OriginalName    string          `json:"original_name"`
	Tags            []string        `json:"tags" gorm:"serializer:json"`
	FileURL         string          `json:"file_url"`
	FileSize        int64           `json:"file_size"`
	Width           int             `json:"width"`
	Height          int             `json:"height"`
	ThumbnailURL    string          `json:"thumbnail_url"`
	ThumbnailSize   int64           `json:"thumbnail_size"`
	ThumbnailWidth  int             `json:"thumbnail_width"`
	ThumbnailHeight int             `json:"thumbnail_height"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	MimeType        string          `json:"mime_type"`
	UserID          uint            `json:"user_id"`
	Albums          []AlbumResponse `json:"albums"`
	StorageName     string          `json:"storage_name"`
}

type GetImagesResponse struct {
	Images   []ImageResponse `json:"images"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// 根据用户ID获取存储实例
func (s *ImageService) getStorageByUserID(userID uint) (storage.Storage, string, error) {
	var user models.User
	result := s.db.Preload("Role").Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, "", cerrors.ErrUserNotFound
	}

	// 根据角色的 StorageName 获取存储配置
	var storageConfig models.Storage
	storageResult := s.db.Where("name = ?", user.Role.StorageName).First(&storageConfig)
	if storageResult.Error != nil {
		// 存储配置不存在，使用默认本地存储
		return storage.NewStorageByStorageName(nil, &s.cfg.Server), user.Role.StorageName, nil
	}

	return storage.NewStorageByStorageName(&storageConfig, &s.cfg.Server), user.Role.StorageName, nil
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

func (s *ImageService) UploadImage(currentUserID uint, AlbumIDs []uint, tags []string, files []*multipart.FileHeader) error {
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	maxThumbSize := s.cfg.SystemSettings.General.MaxThumbSize

	// 获取用户对应的存储实例
	storageInstance, storageName, err := s.getStorageByUserID(currentUserID)
	if err != nil {
		return err
	}

	// 处理每个文件
	for _, file := range files {
		fileUUID := uuid.New().String()
		fileExt := strings.ToLower(filepath.Ext(file.Filename))
		fileOriginalName := file.Filename[:len(file.Filename)-len(fileExt)]
		fileName := fmt.Sprintf("%s$-$%s%s", fileUUID, fileOriginalName, fileExt)
		fileSize := file.Size

		// 执行单个文件上传
		if err := s.executeUpload(storageInstance, storageName, currentUserID, AlbumIDs, tags, file, fileName, fileSize, fileExt, dateDir, maxThumbSize, fileUUID); err != nil {
			log.Errorf("Failed to upload file %s: %v,currentUserID:%d", file.Filename, err, currentUserID)
			return err
		}
	}
	return nil
}

func (s *ImageService) executeUpload(storageInstance storage.Storage, storageName string, currentUserID uint, AlbumIDs []uint, tags []string, file *multipart.FileHeader, fileName string, fileSize int64, fileExt, dateDir string, maxThumbSize uint, fileUUID string) error {
	// 验证所有相册是否存在且属于当前用户
	var albums []models.Album
	if len(AlbumIDs) > 0 {
		result := s.db.Where("id IN ? AND user_id = ?", AlbumIDs, currentUserID).Find(&albums)
		if result.Error != nil {
			log.Errorf("failed to find albums: error=%v", result.Error)
			return cerrors.ErrInternalServer
		}
		if len(albums) != len(AlbumIDs) {
			return cerrors.ErrAlbumNotFound
		}
	}

	mimeType, width, height, err := GetImageDimensions(file)
	if err != nil {
		return err
	}

	// 使用普通上传方法
	fileURL, err := storageInstance.UploadFile(file, "", dateDir, fileName)
	if err != nil {
		return err
	}

	// 生成缩略图，如果失败则清理已上传的文件
	thumbnailURL, thumbnailWidth, thumbnailHeight, thumbnailSize, err := GetThumbnails(dateDir, fileUUID, maxThumbSize, mimeType, file, storageInstance)
	if err != nil {
		// 清理已上传的原始文件
		if deleteErr := storageInstance.DeleteFile(fileURL); deleteErr != nil {
			log.Errorf("failed to delete uploaded file after thumbnail generation failed: %v", deleteErr)
		}
		return err
	}

	// 使用事务处理数据库操作
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	imageModel := &models.Image{
		FileName:        fileName,
		OriginalName:    file.Filename[:len(file.Filename)-len(fileExt)],
		Tags:            tags,
		FileURL:         fileURL,
		FileSize:        fileSize,
		Width:           width,
		Height:          height,
		MimeType:        mimeType,
		UserID:          currentUserID,
		ThumbnailURL:    thumbnailURL,
		ThumbnailSize:   thumbnailSize,
		ThumbnailWidth:  thumbnailWidth,
		ThumbnailHeight: thumbnailHeight,
		StorageName:     storageName,
	}

	result := tx.Create(&imageModel)
	if result.Error != nil {
		tx.Rollback()
		log.Errorf("failed to create image: error=%v", result.Error)
		// 清理已上传的文件和缩略图
		if deleteErr := storageInstance.DeleteFile(fileURL); deleteErr != nil {
			log.Errorf("failed to delete uploaded file after db error: %v", deleteErr)
		}
		if deleteErr := storageInstance.DeleteFile(thumbnailURL); deleteErr != nil {
			log.Errorf("failed to delete thumbnail after db error: %v", deleteErr)
		}
		return cerrors.ErrInternalServer
	}

	// 添加图片到多个相册
	if len(albums) > 0 {
		if err := tx.Model(&imageModel).Association("Albums").Append(&albums); err != nil {
			tx.Rollback()
			log.Errorf("failed to associate albums: error=%v", err)
			// 清理已上传的文件和缩略图
			if deleteErr := storageInstance.DeleteFile(fileURL); deleteErr != nil {
				log.Errorf("failed to delete uploaded file after album association error: %v", deleteErr)
			}
			if deleteErr := storageInstance.DeleteFile(thumbnailURL); deleteErr != nil {
				log.Errorf("failed to delete thumbnail after album association error: %v", deleteErr)
			}
			return cerrors.ErrInternalServer
		}
		// 更新每个相册的图片计数
		for _, album := range albums {
			if err := tx.Model(&album).Update("image_count", album.ImageCount+1).Error; err != nil {
				tx.Rollback()
				log.Errorf("failed to update album image count: error=%v", err)
				// 清理已上传的文件和缩略图
				if deleteErr := storageInstance.DeleteFile(fileURL); deleteErr != nil {
					log.Errorf("failed to delete uploaded file after album update error: %v", deleteErr)
				}
				if deleteErr := storageInstance.DeleteFile(thumbnailURL); deleteErr != nil {
					log.Errorf("failed to delete thumbnail after album update error: %v", deleteErr)
				}
				return cerrors.ErrInternalServer
			}
		}
	}

	// 更新用户的存储空间使用情况
	result = tx.Model(&models.User{}).Where("id = ?", currentUserID).
		Updates(map[string]interface{}{
			"total_size":  gorm.Expr("total_size + ?", fileSize+thumbnailSize),
			"image_count": gorm.Expr("image_count + ?", 1),
		})
	if result.Error != nil {
		tx.Rollback()
		log.Errorf("failed to update user storage usage: id=%d, error=%v", currentUserID, result.Error)
		// 清理已上传的文件和缩略图
		if deleteErr := storageInstance.DeleteFile(fileURL); deleteErr != nil {
			log.Errorf("failed to delete uploaded file after user storage update error: %v", deleteErr)
		}
		if deleteErr := storageInstance.DeleteFile(thumbnailURL); deleteErr != nil {
			log.Errorf("failed to delete thumbnail after user storage update error: %v", deleteErr)
		}
		return cerrors.ErrInternalServer
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		// 清理已上传的文件和缩略图
		if deleteErr := storageInstance.DeleteFile(fileURL); deleteErr != nil {
			log.Errorf("failed to delete uploaded file after transaction commit error: %v", deleteErr)
		}
		if deleteErr := storageInstance.DeleteFile(thumbnailURL); deleteErr != nil {
			log.Errorf("failed to delete thumbnail after transaction commit error: %v", deleteErr)
		}
		return cerrors.ErrInternalServer
	}

	s.db.Preload("Albums").First(&imageModel)

	return nil
}

func (s *ImageService) GetImages(currentUserID uint, page int, pageSize int) (*GetImagesResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var images []ImageResponse
	var imageModels []models.Image
	var total int64

	db := s.db.Model(&imageModels).Where("user_id = ?", currentUserID)
	db.Count(&total)
	db.Preload("Albums").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&imageModels)

	images = MakeImagesWithAlbum(imageModels)

	return &GetImagesResponse{
		Images:   images,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ImageService) GetImage(currentUserID uint, imageID uint) (*ImageResponse, error) {
	var imageModel models.Image
	result := s.db.Preload("Albums").Where("id = ? AND user_id = ?", imageID, currentUserID).First(&imageModel)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrImageNotFound
	}

	imageResponse := MakeImageWithAlbum(imageModel)

	return &imageResponse, nil
}

func (s *ImageService) UpdateImage(currentUserID uint, imageID uint, originalName string, tags []string) (*ImageResponse, error) {
	var imageModel models.Image
	result := s.db.Preload("Albums").Where("id = ? AND user_id = ?", imageID, currentUserID).First(&imageModel)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrImageNotFound
	}

	imageModel.OriginalName = originalName
	imageModel.Tags = tags

	result = s.db.Save(&imageModel)
	if result.Error != nil {
		log.Errorf("failed to update image: id=%d, error=%v", imageID, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	imageResponse := MakeImageWithAlbum(imageModel)

	return &imageResponse, nil
}

func (s *ImageService) DeleteImage(currentUserID uint, imageID uint) error {
	// 使用事务处理删除操作
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var image models.Image
	result := tx.Where("id = ? AND user_id = ?", imageID, currentUserID).First(&image)
	if result.RowsAffected == 0 {
		tx.Rollback()
		return cerrors.ErrImageNotFound
	}

	// Get all albums that the image belongs to
	var albums []models.Album
	tx.Model(&image).Association("Albums").Find(&albums)

	// Remove the image from each album
	tx.Model(&image).Association("Albums").Clear()

	// Update image_count for each album
	for _, album := range albums {
		if err := tx.Model(&album).Update("image_count", album.ImageCount-1).Error; err != nil {
			tx.Rollback()
			log.Errorf("failed to update album image count: error=%v", err)
			return cerrors.ErrInternalServer
		}
	}

	// 更新用户的存储空间使用情况
	result = tx.Model(&models.User{}).Where("id = ?", currentUserID).
		Updates(map[string]interface{}{
			"total_size":  gorm.Expr("total_size - ?", image.FileSize+image.ThumbnailSize),
			"image_count": gorm.Expr("image_count - ?", 1),
		})
	if result.Error != nil {
		tx.Rollback()
		log.Errorf("failed to update user storage usage: id=%d, error=%v", currentUserID, result.Error)
		return cerrors.ErrInternalServer
	}

	// 删除数据库中的图片记录
	result = tx.Delete(&image)
	if result.Error != nil {
		tx.Rollback()
		log.Errorf("failed to delete image: id=%d, error=%v", imageID, result.Error)
		return cerrors.ErrInternalServer
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		return cerrors.ErrInternalServer
	}

	// 获取图片对应的存储实例
	storageInstance, err := s.getStorageByStorageName(image.StorageName)
	if err != nil {
		return err
	}

	// Delete the image from storage
	if err := storageInstance.DeleteFile(image.FileURL); err != nil {
		log.Errorf("failed to delete image file: id=%d, error=%v", imageID, err)
		return cerrors.ErrInternalServer
	}

	if err := storageInstance.DeleteFile(image.ThumbnailURL); err != nil {
		log.Errorf("failed to delete thumbnail image: id=%d, error=%v", imageID, err)
		return cerrors.ErrInternalServer
	}

	return nil
}

func (s *ImageService) AddImageToAlbum(currentUserID uint, imageID uint, albumID uint) error {
	var album models.Album
	result := s.db.Where("id = ? AND user_id = ?", albumID, currentUserID).First(&album)
	if result.RowsAffected == 0 {
		return cerrors.ErrAlbumNotFound
	}

	var image models.Image
	result = s.db.Where("id = ? AND user_id = ?", imageID, currentUserID).First(&image)
	if result.RowsAffected == 0 {
		return cerrors.ErrImageNotFound
	}

	// Judge if the image is already in the album
	var count int64
	s.db.Model(&models.ImageAlbum{}).
		Where("image_id = ? AND album_id = ?", imageID, albumID).
		Count(&count)
	if count > 0 {
		return cerrors.ErrImageAlreadyInAlbum
	}

	s.db.Model(&image).Association("Albums").Append(&album)
	s.db.Model(&album).Update("image_count", album.ImageCount+1)

	return nil
}

func (s *ImageService) RemoveImageFromAlbum(currentUserID uint, imageID uint, albumID uint) error {
	var album models.Album
	result := s.db.Where("id = ? AND user_id = ?", albumID, currentUserID).First(&album)
	if result.RowsAffected == 0 {
		return cerrors.ErrAlbumNotFound
	}

	var image models.Image
	result = s.db.Where("id = ? AND user_id = ?", imageID, currentUserID).First(&image)
	if result.RowsAffected == 0 {
		return cerrors.ErrImageNotFound
	}

	// Judge if the image is in the album
	var count int64
	s.db.Model(&models.ImageAlbum{}).
		Where("image_id = ? AND album_id = ?", imageID, albumID).
		Count(&count)
	if count == 0 {
		return cerrors.ErrImageNotInAlbum
	}

	s.db.Model(&image).Association("Albums").Delete(&album)
	s.db.Model(&album).Update("image_count", album.ImageCount-1)

	return nil
}

func (s *ImageService) SearchImagesByTagsOrTitle(currentUserID uint, searchKey string, page int, pageSize int) (*GetImagesResponse, error) {
	offset := (page - 1) * pageSize

	var imageModels []models.Image
	var images []ImageResponse
	var total int64
	var db *gorm.DB

	if searchKey != "" {
		sk := fmt.Sprintf("%%%s%%", searchKey)
		db = s.db.Model(&imageModels).Where("user_id = ? AND (tags LIKE ? OR original_name LIKE ?)", currentUserID, sk, sk)
	}

	db.Count(&total)
	db.Preload("Albums").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&imageModels)

	images = MakeImagesWithAlbum(imageModels)

	return &GetImagesResponse{
		Images:   images,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ImageService) UploadImageLimitCheck(currentUserID uint, files []*multipart.FileHeader) error {
	var user models.User
	result := s.db.Preload("Role").Where("id = ?", currentUserID).First(&user)
	if result.RowsAffected == 0 {
		return cerrors.ErrUserNotFound
	}
	allowedExtensions := user.Role.AllowedExtensions
	uploadFileCount := len(files)
	uploadFileSize := 0

	// -1 means no limit
	if user.Role.MaxFilesPerUpload != -1 {
		if uploadFileCount > user.Role.MaxFilesPerUpload {
			return cerrors.ErrMaxFilesPerUpload
		}
	}

	for _, file := range files {
		fileExt, _, err := GetFileType(file)
		if err != nil {
			return err
		}
		if !slices.Contains(allowedExtensions, "."+fileExt) {
			return cerrors.ErrAllowedExtensions
		}
		if user.Role.MaxFileSizeMB != -1 {
			if file.Size > int64(user.Role.MaxFileSizeMB)*1024*1024 {
				return cerrors.ErrMaxFileSizeMB
			}
		}
		uploadFileSize += int(file.Size)
	}

	if user.Role.MaxStorageSizeMB != -1 {
		StorageUsed := user.TotalSize
		if StorageUsed+int64(uploadFileSize) > int64(user.Role.MaxStorageSizeMB)*1024*1024 {
			return cerrors.ErrMaxStorageSizeMB
		}
	}

	return nil
}

func GetThumbnails(dateDir, fileUUID string, maxThumbSize uint, MimeType string, File *multipart.FileHeader, ostorage storage.Storage) (string, int, int, int64, error) {
	// 根据 MimeType 确定缩略图扩展名
	thumbnailExt := "jpg"
	if MimeType == "image/gif" {
		thumbnailExt = "gif"
	}
	thumbnailName := fmt.Sprintf("%s_thumbnail.%s", fileUUID, thumbnailExt)
	err := os.MkdirAll("tmp", 0755)
	if err != nil {
		log.Errorf("failed to create temporary directory: path=%s, error=%v", "tmp", err)
		return "", 0, 0, 0, cerrors.ErrInternalServer
	}
	thumbnailPath := filepath.Join("tmp", thumbnailName)

	file, err := File.Open()
	if err != nil {
		log.Errorf("failed to open file: path=%s, error=%v", File.Filename, err)
		return "", 0, 0, 0, cerrors.ErrInternalServer
	}
	defer file.Close()

	// 处理 GIF 文件的动画
	if MimeType == "image/gif" {
		// 重置文件指针到开头
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			log.Errorf("failed to seek file: error=%v", err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}

		// 解码整个 GIF 动画
		gifImg, err := gif.DecodeAll(file)
		if err != nil {
			log.Errorf("failed to decode gif: error=%v", err)
			return "", 0, 0, 0, cerrors.ErrDecodeImage
		}

		// 处理每一帧
		for i, frame := range gifImg.Image {
			// 缩放每一帧
			resizedFrame := resize.Thumbnail(maxThumbSize, maxThumbSize, frame, resize.Lanczos3)
			// 将image.Image转换为*image.Paletted
			if palettedFrame, ok := resizedFrame.(*image.Paletted); ok {
				gifImg.Image[i] = palettedFrame
			} else {
				// 如果转换失败，创建一个新的Paletted图像
				bounds := resizedFrame.Bounds()
				palettedFrame := image.NewPaletted(bounds, frame.Palette)
				draw.Draw(palettedFrame, bounds, resizedFrame, bounds.Min, draw.Src)
				gifImg.Image[i] = palettedFrame
			}
		}

		// 更新 GIF 配置的尺寸为第一帧的尺寸
		if len(gifImg.Image) > 0 {
			firstFrame := gifImg.Image[0]
			gifImg.Config.Width = firstFrame.Bounds().Dx()
			gifImg.Config.Height = firstFrame.Bounds().Dy()
		}

		// 创建临时文件
		tempFile, err := os.Create(thumbnailPath)
		if err != nil {
			log.Errorf("failed to create temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}

		// 编码整个动画
		if err := gif.EncodeAll(tempFile, gifImg); err != nil {
			tempFile.Close()
			if err := os.Remove(thumbnailPath); err != nil {
				log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
				return "", 0, 0, 0, cerrors.ErrInternalServer
			}
			log.Errorf("failed to encode gif thumbnail: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrEncodeImage
		}

		// 关闭文件
		if err := tempFile.Close(); err != nil {
			if err := os.Remove(thumbnailPath); err != nil {
				log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
				return "", 0, 0, 0, cerrors.ErrInternalServer
			}
			log.Errorf("failed to close temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}

		// 获取文件信息
		fileInfo, err := os.Stat(thumbnailPath)
		if err != nil {
			log.Errorf("failed to get thumbnail file info: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}

		thumbnailSize := fileInfo.Size()
		// 使用第一帧的尺寸作为缩略图尺寸
		if len(gifImg.Image) > 0 {
			firstFrame := gifImg.Image[0]
			thumbnailWidth := firstFrame.Bounds().Dx()
			thumbnailHeight := firstFrame.Bounds().Dy()

			// 上传文件
			thumbnailURL, err := ostorage.UploadFile(nil, thumbnailPath, dateDir, thumbnailName)
			if err != nil {
				if err := os.Remove(thumbnailPath); err != nil {
					log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
					return "", 0, 0, 0, cerrors.ErrInternalServer
				}
				return "", 0, 0, 0, err
			}

			// 清理临时文件
			if err := os.Remove(thumbnailPath); err != nil {
				log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
				return "", 0, 0, 0, cerrors.ErrInternalServer
			}

			return thumbnailURL, thumbnailWidth, thumbnailHeight, thumbnailSize, nil
		}

		// 如果没有帧，返回错误
		if err := os.Remove(thumbnailPath); err != nil {
			log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
		}
		return "", 0, 0, 0, cerrors.ErrDecodeImage
	} else {
		// 处理其他文件类型
		var img image.Image
		switch MimeType {
		case "image/bmp":
			img, err = bmp.Decode(file)
		case "image/tiff", "image/tif":
			img, err = tiff.Decode(file)
		case "image/webp":
			img, err = webp.Decode(file)
		case "image/svg+xml":
			return "", 0, 0, 0, cerrors.ErrDecodeImage
		default:
			// 对于 jpeg, png，image.Decode 会自动处理
			img, _, err = image.Decode(file)
		}

		if err != nil {
			return "", 0, 0, 0, cerrors.ErrDecodeImage
		}

		canvas := resize.Thumbnail(maxThumbSize, maxThumbSize, img, resize.Lanczos3)

		// 创建临时文件
		tempFile, err := os.Create(thumbnailPath)
		if err != nil {
			log.Errorf("failed to create temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}

		// 编码为 JPG
		if err := jpeg.Encode(tempFile, canvas, &jpeg.Options{Quality: 85}); err != nil {
			tempFile.Close()
			if err := os.Remove(thumbnailPath); err != nil {
				log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
				return "", 0, 0, 0, cerrors.ErrInternalServer
			}
			log.Errorf("failed to encode thumbnail image: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrEncodeImage
		}

		// 关闭文件
		if err := tempFile.Close(); err != nil {
			if err := os.Remove(thumbnailPath); err != nil {
				log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
				return "", 0, 0, 0, cerrors.ErrInternalServer
			}
			log.Errorf("failed to close temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}

		// 获取文件的实际大小
		fileInfo, err := os.Stat(thumbnailPath)
		if err != nil {
			log.Errorf("failed to get thumbnail file info: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}
		thumbnailSize := fileInfo.Size()
		thumbnailWidth := canvas.Bounds().Dx()
		thumbnailHeight := canvas.Bounds().Dy()

		thumbnailURL, err := ostorage.UploadFile(nil, thumbnailPath, dateDir, thumbnailName)
		if err != nil {
			if err := os.Remove(thumbnailPath); err != nil {
				log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
				return "", 0, 0, 0, cerrors.ErrInternalServer
			}
			return "", 0, 0, 0, err
		}

		// 清理临时文件
		if err := os.Remove(thumbnailPath); err != nil {
			log.Errorf("failed to remove temporary thumbnail file: path=%s, error=%v", thumbnailPath, err)
			return "", 0, 0, 0, cerrors.ErrInternalServer
		}

		return thumbnailURL, thumbnailWidth, thumbnailHeight, thumbnailSize, nil
	}
}

func GetImageDimensions(File *multipart.FileHeader) (string, int, int, error) {
	file, err := File.Open()
	if err != nil {
		log.Errorf("failed to open file: error=%v", err)
		return "", 0, 0, cerrors.ErrInternalServer
	}
	defer file.Close()

	fileExt, mimeType, err := GetFileType(File)
	if err != nil {
		return "", 0, 0, err
	}

	var img image.Config

	switch fileExt {
	case "bmp":
		img, err = bmp.DecodeConfig(file)
	case "tiff", "tif":
		img, err = tiff.DecodeConfig(file)
	case "webp":
		img, err = webp.DecodeConfig(file)
	default:
		img, _, err = image.DecodeConfig(file)
	}

	if err != nil {
		return "", 0, 0, cerrors.ErrDecodeImage
	}

	return mimeType, img.Width, img.Height, nil
}

func GetFileType(File *multipart.FileHeader) (string, string, error) {
	file, err := File.Open()
	if err != nil {
		log.Errorf("failed to open file: error=%v", err)
		return "", "", cerrors.ErrInternalServer
	}
	defer file.Close()

	// 读取文件前1024字节用于文件类型检测
	buffer := make([]byte, 1024)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Errorf("failed to read file: error=%v", err)
		return "", "", cerrors.ErrInternalServer
	}

	// 匹配文件类型
	kind, _ := filetype.Match(buffer)
	if kind == filetype.Unknown {
		return "", "", cerrors.ErrUnknownFileType
	}
	return kind.Extension, kind.MIME.Value, nil
}

func MakeImagesWithAlbum(imageModels []models.Image) []ImageResponse {
	images := make([]ImageResponse, 0) // 非 nil 空切片

	for _, imageModel := range imageModels {
		albumResponses := make([]AlbumResponse, len(imageModel.Albums))
		for i, album := range imageModel.Albums {
			albumResponses[i] = AlbumResponse{
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
			}
		}
		images = append(images, ImageResponse{
			ID:              imageModel.ID,
			FileName:        imageModel.FileName,
			OriginalName:    imageModel.OriginalName,
			Tags:            imageModel.Tags,
			FileURL:         imageModel.FileURL,
			FileSize:        imageModel.FileSize,
			Width:           imageModel.Width,
			Height:          imageModel.Height,
			ThumbnailURL:    imageModel.ThumbnailURL,
			ThumbnailSize:   imageModel.ThumbnailSize,
			ThumbnailWidth:  imageModel.ThumbnailWidth,
			ThumbnailHeight: imageModel.ThumbnailHeight,
			CreatedAt:       imageModel.CreatedAt,
			UpdatedAt:       imageModel.UpdatedAt,
			MimeType:        imageModel.MimeType,
			UserID:          imageModel.UserID,
			Albums:          albumResponses,
			StorageName:     imageModel.StorageName,
		})
	}

	return images
}

func MakeImageWithAlbum(imageModel models.Image) ImageResponse {
	albumResponse := make([]AlbumResponse, len(imageModel.Albums))
	for i, album := range imageModel.Albums {
		albumResponse[i] = AlbumResponse{
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
		}
	}
	return ImageResponse{
		ID:              imageModel.ID,
		FileName:        imageModel.FileName,
		OriginalName:    imageModel.OriginalName,
		Tags:            imageModel.Tags,
		FileURL:         imageModel.FileURL,
		FileSize:        imageModel.FileSize,
		Width:           imageModel.Width,
		Height:          imageModel.Height,
		ThumbnailURL:    imageModel.ThumbnailURL,
		ThumbnailSize:   imageModel.ThumbnailSize,
		ThumbnailWidth:  imageModel.ThumbnailWidth,
		ThumbnailHeight: imageModel.ThumbnailHeight,
		CreatedAt:       imageModel.CreatedAt,
		UpdatedAt:       imageModel.UpdatedAt,
		MimeType:        imageModel.MimeType,
		UserID:          imageModel.UserID,
		Albums:          albumResponse,
		StorageName:     imageModel.StorageName,
	}
}
