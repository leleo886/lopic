package models

type ImageAlbum struct {
	ImageID uint `gorm:"primaryKey" json:"image_id"`
	AlbumID uint `gorm:"primaryKey" json:"album_id"`
	Image   Image `gorm:"foreignKey:ImageID" json:"image"`
	Album   Album `gorm:"foreignKey:AlbumID" json:"album"`
}

func (ImageAlbum) TableName() string {
	return "image_albums"
}
