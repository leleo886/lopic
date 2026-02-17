package admin_services

import (
	"fmt"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"gorm.io/gorm"
)

type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

type RoleRequest struct {
	Name              string                     `json:"name"`
	Description       string                     `json:"description"`
	AllowedExtensions []string                   `json:"allowed_extensions"`
	MaxFilesPerUpload int                        `json:"max_files_per_upload"`
	MaxFileSizeMB     int                        `json:"max_file_size_mb"`
	MaxAlbumsPerUser  int                        `json:"max_albums_per_user"`
	MaxStorageSizeMB  int                        `json:"max_storage_size_mb"`
	GalleryOpen       bool                       `json:"gallery_open"`
	StorageName       string                     `json:"storage_name"`
}

func (s *RoleService) GetRoles(page, pageSize, offset int, searchkey, orderby, order string) (*[]models.Role, error) {
	var roles []models.Role
	query := s.db
	if searchkey != "" {
		sk := fmt.Sprintf("%%%s%%", searchkey)
		query = query.Where("name LIKE ? OR description LIKE ? OR allowed_extensions LIKE ?", sk, sk, sk)
	}
	result := query.Limit(pageSize).Offset(offset).
		Order(fmt.Sprintf("%s %s", orderby, order)).Find(&roles)
	if result.Error != nil {
		log.Errorf("failed to get roles: error=%v", result.Error)
		return nil, cerrors.ErrInternalServer
	}
	return &roles, nil
}

func (s *RoleService) GetRole(id int) (*models.Role, error) {
	var role models.Role
	result := s.db.Where("id = ?", id).First(&role)
	if result.Error != nil {
		return nil, cerrors.ErrRoleNotFound
	}
	return &role, nil
}

func (s *RoleService) CreateRole(role *models.Role) error {
	// 先检查 name 是否已存在
	var existingRole models.Role
	result := s.db.Where("name = ?", role.Name).First(&existingRole)
	if result.RowsAffected > 0 {
		return cerrors.ErrRoleNameExists
	}

	// 检查 storage_name 是否存在
	var existingStorage models.Storage
	result = s.db.Where("name = ?", role.StorageName).First(&existingStorage)
	if result.RowsAffected == 0 {
		return cerrors.ErrStorageNotFound
	}

	result = s.db.Create(role)
	if result.Error != nil {
		log.Errorf("failed to create role: error=%v", result.Error)
		return cerrors.ErrInternalServer
	}
	return nil
}

func (s *RoleService) UpdateRole(id uint, role *RoleRequest) error {
	var existingRole models.Role
	result := s.db.Where("id = ?", id).First(&existingRole)
	if result.Error != nil {
		return cerrors.ErrRoleNotFound
	}

	var existingStorage models.Storage
	result = s.db.Where("name = ?", role.StorageName).First(&existingStorage)
	if result.RowsAffected == 0 {
		return cerrors.ErrStorageNotFound
	}

	if existingRole.Name == "admin" && role.Name != "admin" {
		return cerrors.ErrCannotChangeAdminRoleName
	}
	if existingRole.Name == "user" && role.Name != "user" {
		return cerrors.ErrCannotChangeUserRoleName
	}

	existingRole.Name = role.Name
	existingRole.Description = role.Description
	existingRole.AllowedExtensions = role.AllowedExtensions
	existingRole.MaxFilesPerUpload = role.MaxFilesPerUpload
	existingRole.MaxFileSizeMB = role.MaxFileSizeMB
	existingRole.MaxAlbumsPerUser = role.MaxAlbumsPerUser
	existingRole.MaxStorageSizeMB = role.MaxStorageSizeMB
	existingRole.GalleryOpen = role.GalleryOpen
	existingRole.StorageName = role.StorageName

	result = s.db.Save(&existingRole)
	if result.Error != nil {
		log.Errorf("failed to update role: id=%d, error=%v", id, result.Error)
		return cerrors.ErrInternalServer
	}
	return nil
}

func (s *RoleService) DeleteRole(id uint) error {
	var existingRole models.Role
	result := s.db.Where("id = ?", id).First(&existingRole)
	if result.Error != nil {
		return cerrors.ErrRoleNotFound
	}

	if existingRole.Name == "admin" || existingRole.Name == "user" {
		return cerrors.ErrCannotDeleteDefaultRole
	}

	// 先查找是否有用户关联该角色
	var users []models.User
	result = s.db.Where("role_id = ?", id).Find(&users)
	if result.RowsAffected > 0 {
		return cerrors.ErrRoleAssociatedWithUsers
	}

	// 删除角色
	result = s.db.Delete(&existingRole)
	if result.Error != nil {
		log.Errorf("failed to delete role: id=%d, error=%v", id, result.Error)
		return cerrors.ErrInternalServer
	}
	return nil
}

func (s *RoleService) GetUsersCountByRole() (map[string]int, error) {
	var users []models.User
	result := s.db.Model(&models.User{}).Preload("Role").Find(&users)
	if result.Error != nil {
		log.Errorf("failed to get users count by role: error=%v", result.Error)
		return nil, cerrors.ErrInternalServer
	}

	counts := make(map[string]int)
	for _, user := range users {
		if user.Role.Name != "" {
			counts[user.Role.Name]++
		}
	}
	return counts, nil
}
