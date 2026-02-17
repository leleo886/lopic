package models

type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password string `gorm:"size:100;not null" json:"-"`
	Email    string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	RoleID   uint   `gorm:"not null;default:2" json:"role_id"`
	Role     Role   `gorm:"foreignKey:RoleID" json:"role"`
	Active   bool   `gorm:"default:false" json:"active"`
	TotalSize   int64 `gorm:"not null;default:0" json:"total_size"`
    ImageCount  int   `gorm:"not null;default:0" json:"image_count"`
}


func (User) TableName() string {
	return "users"
}
