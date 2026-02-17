package models

type Image struct {
	BaseModel
	FileName        string   `gorm:"size:255;not null" json:"file_name"`
	OriginalName    string   `gorm:"size:255;not null" json:"original_name"`
	FileURL         string   `gorm:"size:500;not null;uniqueIndex" json:"file_url"`
	FileSize        int64    `gorm:"not null" json:"file_size"`
	Width           int      `gorm:"not null" json:"width"`
	Height          int      `gorm:"not null" json:"height"`
	MimeType        string   `gorm:"size:50;not null" json:"mime_type"`
	UserID          uint     `gorm:"not null;index" json:"user_id"`
	User            User     `gorm:"foreignKey:UserID" json:"user"`
	Albums          []Album  `gorm:"many2many:image_albums;" json:"albums"`
	ThumbnailURL    string   `gorm:"size:500" json:"thumbnail_url"`
	ThumbnailSize   int64    `gorm:"not null" json:"thumbnail_size"`
	ThumbnailWidth  int      `gorm:"not null" json:"thumbnail_width"`
	ThumbnailHeight int      `gorm:"not null" json:"thumbnail_height"`
	Tags            []string `gorm:"type:json;serializer:json;" json:"tags"`
	StorageName     string   `gorm:"size:50;not null;default:'local'" json:"storage_name"` // 存储配置名称
}

func (Image) TableName() string {
	return "images"
}
