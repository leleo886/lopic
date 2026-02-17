package services

import (
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"sort"
)

type UserService struct {
	db *gorm.DB
	cfg *config.Config
}

func NewUserService(db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{db: db, cfg: cfg}
}

type StorageUsage struct {
	TotalSize  int64 `json:"total_size"`
	ImageCount int   `json:"image_count"`
}

type TagCloudItem struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

func (s *UserService) GetMe(currentUserID uint) (*models.User, error) {
	var user models.User
	result := s.db.Preload("Role").First(&user, currentUserID)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrUserNotFound
	}
	return &user, nil
}

func (s *UserService) UpdateMe(currentUserID uint, username, password string) (*models.User, error) {
	// 获取要更新的用户信息
	var user models.User
	result := s.db.Preload("Role").First(&user, currentUserID)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrUserNotFound
	}

	// 更新用户信息
	user.Username = username

	// 如果密码不为空，则更新密码
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, cerrors.ErrPwdEncFailed
		}
		user.Password = string(hashedPassword)
	}

	result = s.db.Save(&user)
	if result.Error != nil {
		log.Errorf("failed to update user: id=%d, error=%v", currentUserID, result.Error)
		return nil, cerrors.ErrInternalServer
	}
	return &user, nil
}

// GetStorageUsage 获取用户存储空间使用情况
func (s *UserService) GetStorageUsage(userID uint) (*StorageUsage, error) {
	var user models.User
	result := s.db.First(&user, userID)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrUserNotFound
	}
	return &StorageUsage{
		TotalSize:  user.TotalSize,
		ImageCount: user.ImageCount,
	}, nil
}

// GetImagesTagsCloud 获取用户图片标签云
func (s *UserService) GetImagesTagsCloud(userID uint) ([]TagCloudItem, error) {
	// 1. 找到所有用户上传的图片的标签数组
	var images []models.Image
	result := s.db.Model(&models.Image{}).
		Where("user_id = ? AND tags IS NOT NULL", userID).
		Find(&images)

	if result.Error != nil {
		return nil, result.Error
	}
	// 2. 统计标签出现次数
	tagsCount := utils.WordCount(images)

	// 3. 转换为 TagCloudItem 切片
	var tagCloudItems []TagCloudItem
	for tag, count := range tagsCount {
		tagCloudItems = append(tagCloudItems, TagCloudItem{
			Tag:   tag,
			Count: count,
		})
	}

	// 按数量降序排序
	sort.Slice(tagCloudItems, func(i, j int) bool {
		return tagCloudItems[i].Count > tagCloudItems[j].Count
	})

	maxTags := s.cfg.SystemSettings.General.MaxTags

	//不管多少个标签，都返回前maxTags个
	if len(tagCloudItems) > maxTags {
		tagCloudItems = tagCloudItems[:maxTags]
	}

	return tagCloudItems, nil
}
