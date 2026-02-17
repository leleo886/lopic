package models

import "time"

type PasswordResetCode struct {
	BaseModel
	Email     string    `gorm:"size:100;not null;index" json:"email"`
	Code      string    `gorm:"size:6;not null" json:"code"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
}

func (PasswordResetCode) TableName() string {
	return "password_reset_codes"
}
