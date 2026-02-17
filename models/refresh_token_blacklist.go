package models

import (
	"time"

)

type RefreshTokenBlacklist struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TokenHash string    `gorm:"uniqueIndex;size:64;not null" json:"token_hash"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RefreshTokenBlacklist) TableName() string {
	return "refresh_token_blacklist"
}

