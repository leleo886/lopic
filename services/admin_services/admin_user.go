package admin_services

import (
	"fmt"
	"sort"

	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/storage"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/services"
	"github.com/leleo886/lopic/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewUserService(db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{db: db, cfg: cfg}
}

// UserRequest 用户管理请求结构体
type UserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required"`
	Active   bool   `json:"active"`
}

type GetUsersResponse struct {
	Users    []models.User `json:"users"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Total    int64         `json:"total"`
}

func (s *UserService) GetUsers(page, pageSize, offset int, searchkey, orderby, order string) (*GetUsersResponse, error) {
	var users []models.User
	query := s.db.Preload("Role")
	if searchkey != "" {
		sk := fmt.Sprintf("%%%s%%", searchkey)
		query = query.Where("username LIKE ? OR email LIKE ?", sk, sk)
	}
	result := query.Limit(pageSize).Offset(offset).
		Order(fmt.Sprintf("%s %s", orderby, order)).Find(&users)
	if result.Error != nil {
		log.Errorf("failed to get users: error=%v", result.Error)
		return nil, cerrors.ErrInternalServer
	}
	return &GetUsersResponse{
		Users:    users,
		Page:     page,
		PageSize: pageSize,
		Total:    result.RowsAffected,
	}, nil
}

func (s *UserService) GetUser(id int) (*models.User, error) {
	var user models.User
	result := s.db.Preload("Role").First(&user, id)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrUserNotFound
	}
	return &user, nil
}

func (s *UserService) CreateUser(req UserRequest) error {
	tx := s.db.Begin() // 开启事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查用户名是否已存在
	var existingUser models.User
	result := tx.Where("username = ?", req.Username).First(&existingUser)
	if result.RowsAffected > 0 {
		tx.Rollback()
		return cerrors.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	result = tx.Where("email = ?", req.Email).First(&existingUser)
	if result.RowsAffected > 0 {
		tx.Rollback()
		return cerrors.ErrEmailExists
	}

	// 加密密码
	if len(req.Password) < 6 {
		tx.Rollback()
		return cerrors.ErrBadRequest
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return cerrors.ErrPwdEncFailed
	}

	// 获取角色ID
	var role models.Role
	result = tx.Where("name = ?", req.Role).First(&role)
	if result.RowsAffected == 0 {
		tx.Rollback()
		return cerrors.ErrRoleNotFound
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		RoleID:   role.ID,
		Active:   req.Active,
	}

	result = tx.Create(&user)
	if result.Error != nil {
		tx.Rollback()
		return cerrors.ErrCreateUserFailed
	}

	tx.Commit() // 提交事务

	return nil
}

func (s *UserService) UpdateUser(id int, currentUserID uint, req UserRequest) (*models.User, error) {
	tx := s.db.Begin() // 开启事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取要更新的用户信息
	var user models.User
	result := tx.Preload("Role").First(&user, id)
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, cerrors.ErrUserNotFound
	}
	if user.Role.Name == "admin" && req.Role != "admin" {
		tx.Rollback()
		return nil, cerrors.ErrOneAdmin
	} else if user.Role.Name != "admin" && req.Role == "admin" {
		tx.Rollback()
		return nil, cerrors.ErrOneAdmin
	}

	// 防止管理员禁用自己
	if id == int(currentUserID) && !req.Active {
		tx.Rollback()
		return nil, cerrors.ErrCannotDisableSelf
	}

	// 获取角色ID
	var role models.Role
	result = tx.Where("name = ?", req.Role).First(&role)
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, cerrors.ErrRoleNotFound
	}

	// 更新用户信息
	user.Username = req.Username
	user.Email = req.Email
	user.RoleID = role.ID
	user.Role = role // 同时更新 Role 字段，确保 GORM 正确保存 RoleID 的更新
	user.Active = req.Active

	fmt.Printf("user=%v\n", user.RoleID)

	// 如果密码不为空，则更新密码
	if req.Password != "" {
		if len(req.Password) < 6 {
			tx.Rollback()
			return nil, cerrors.ErrBadRequest
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return nil, cerrors.ErrPwdEncFailed
		}
		user.Password = string(hashedPassword)
	}

	result = tx.Save(&user)
	if result.Error != nil {
		tx.Rollback()
		log.Errorf("failed to update user: id=%d, error=%v", id, result.Error)
		return nil, cerrors.ErrInternalServer
	}

	tx.Commit() // 提交事务

	return &user, nil
}

func (s *UserService) DeleteUser(id int) error {
	// 使用事务处理删除操作
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var user models.User
	result := tx.Preload("Role").First(&user, id)
	if result.RowsAffected == 0 {
		tx.Rollback()
		return cerrors.ErrUserNotFound
	}

	if user.Role.Name == "admin" {
		tx.Rollback()
		return cerrors.ErrCannotDeleteAdminUser
	}

	// 查询用户所有关联的相册
	var albums []models.Album
	var images []models.Image
	result = tx.Where("user_id = ?", id).Find(&albums)
	if result.Error != nil {
		tx.Rollback()
		return cerrors.ErrFailedToDeleteUser
	}

	// 为每个相册清除图片关联
	for _, album := range albums {
		if err := tx.Model(&album).Association("Images").Clear(); err != nil {
			tx.Rollback()
			log.Errorf("failed to clear album images association: album_id=%d, error=%v", album.ID, err)
			return cerrors.ErrFailedToDeleteUser
		}
	}

	// 删除用户所有关联的相册
	result = tx.Where("user_id = ?", id).Delete(&models.Album{})
	if result.Error != nil {
		tx.Rollback()
		return cerrors.ErrFailedToDeleteUser
	}

	// 查询用户的所有图片（在事务内查询，确保一致性）
	result = tx.Where("user_id = ?", id).Find(&images)
	if result.Error != nil {
		tx.Rollback()
		return cerrors.ErrFailedToDeleteUser
	}

	// 删除用户的所有图片记录（在事务内删除，防止并发问题）
	result = tx.Where("user_id = ?", id).Delete(&models.Image{})
	if result.Error != nil {
		tx.Rollback()
		return cerrors.ErrFailedToDeleteUser
	}

	// 删除用户记录
	result = tx.Delete(&user)
	if result.Error != nil {
		tx.Rollback()
		return cerrors.ErrFailedToDeleteUser
	}

	// 提交事务（所有数据库操作完成）
	if err := tx.Commit().Error; err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		return cerrors.ErrFailedToDeleteUser
	}

	// 删除存储中的文件（在事务提交后执行，因为文件系统操作无法回滚）
	// 注意：这里使用 goroutine 异步删除，避免阻塞主流程
	// 如果删除失败，记录日志但不影响删除操作的成功
	go func(images []models.Image) {
		for _, image := range images {
			storageInstance, err := s.getStorageByStorageName(image.StorageName)
			if err != nil {
				log.Errorf("failed to get storage instance: storage_name=%s, error=%v", image.StorageName, err)
				continue
			}

			if err := storageInstance.DeleteFile(image.FileURL); err != nil {
				log.Errorf("failed to delete image file: user_id=%d, image_id=%d, error=%v", image.UserID, image.ID, err)
			}

			if err := storageInstance.DeleteFile(image.ThumbnailURL); err != nil {
				log.Errorf("failed to delete thumbnail file: user_id=%d, image_id=%d, error=%v", image.UserID, image.ID, err)
			}
		}
	}(images)

	return nil
}

// 根据存储名称获取存储实例
func (s *UserService) getStorageByStorageName(storageName string) (storage.Storage, error) {
	var storageConfig models.Storage
	result := s.db.Where("name = ?", storageName).First(&storageConfig)
	if result.Error != nil {
		// 存储配置不存在，使用默认本地存储
		return storage.NewStorageByStorageName(nil, &s.cfg.Server), nil
	}

	return storage.NewStorageByStorageName(&storageConfig, &s.cfg.Server), nil
}

func (s *UserService) GetAllImagesTagsCloud() ([]services.TagCloudItem, error) {
	// 1. 找到所有用户上传的图片的标签数组
	var images []models.Image
	result := s.db.Model(&models.Image{}).
		Where("tags IS NOT NULL AND tags != ''").
		Find(&images)

	if result.Error != nil {
		return nil, result.Error
	}

	// 2. 统计标签出现次数
	tagsCount := utils.WordCount(images)

	// 3. 转换为 TagCloudItem 切片
	var tagCloudItems []services.TagCloudItem
	for tag, count := range tagsCount {
		tagCloudItems = append(tagCloudItems, services.TagCloudItem{
			Tag:   tag,
			Count: count,
		})
	}

	// 按数量降序排序
	sort.Slice(tagCloudItems, func(i, j int) bool {
		return tagCloudItems[i].Count > tagCloudItems[j].Count
	})

	return tagCloudItems, nil
}
