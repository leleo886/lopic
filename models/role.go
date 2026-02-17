package models

import (
	"gorm.io/gorm"
)

// Role -1表示无限制
type Role struct {
	BaseModel
	Name              string   `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Description       string   `gorm:"size:200" json:"description"`
	Users             []User   `gorm:"foreignKey:RoleID" json:"users"`
	AllowedExtensions []string `gorm:"type:json;serializer:json;" json:"allowed_extensions"`
	MaxFilesPerUpload int      `gorm:"default:10" json:"max_files_per_upload"`
	MaxFileSizeMB     int      `gorm:"default:5" json:"max_file_size_mb"`
	MaxAlbumsPerUser  int      `gorm:"default:5" json:"max_albums_per_user"`
	MaxStorageSizeMB  int      `gorm:"default:300" json:"max_storage_size_mb"`
	GalleryOpen       bool     `gorm:"default:false" json:"gallery_open"`
	StorageName       string   `gorm:"size:50;default:'local'" json:"storage_name"` // 存储配置名称
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if len(r.AllowedExtensions) == 0 {
		r.AllowedExtensions = []string{
			".jpg",
			".jpeg",
			".png",
		}
	}

	// 设置默认存储配置名称
	if r.StorageName == "" {
		r.StorageName = "local"
	}
	return nil
}

func (Role) TableName() string {
	return "roles"
}
