package models

type Album struct {
	BaseModel
	Name           string  `gorm:"size:100;not null" json:"name"`
	Description    string  `gorm:"size:500" json:"description"`
	UserID         uint    `gorm:"not null;index" json:"user_id"`
	User           User    `gorm:"foreignKey:UserID" json:"user"`
	CoverImage     string  `gorm:"size:500" json:"cover_image"`
	ImageCount     int     `gorm:"default:0" json:"image_count"`
	Images         []Image `gorm:"many2many:image_albums;" json:"images"`
	GalleryEnabled bool    `gorm:"default:false" json:"gallery_enabled"`
	SerialNumber   int     `gorm:"default:0" json:"serial_number"`
}

func (Album) TableName() string {
	return "albums"
}
