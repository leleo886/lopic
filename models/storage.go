package models

import (
	"gorm.io/gorm"
)

// Storage 存储配置模型
type Storage struct {
	BaseModel
	Name      string        `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Type      string        `gorm:"size:20;not null" json:"type"` // local, webdav
	Config    StorageConfig `gorm:"type:json;serializer:json;not null" json:"config"`
}

// StorageConfig 存储配置结构
type StorageConfig struct {
	// 通用配置
	BasePath  string `json:"base_path"`
	StaticURL string `json:"static_url"`

	// WebDAV 特定配置
	BaseURL  string `json:"base_url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Storage) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (Storage) TableName() string {
	return "storages"
}
