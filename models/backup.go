package models

import (
	"time"
)

type BackupTask struct {
	BaseModel
	Status      string     `gorm:"size:20;not null" json:"status"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Size        int64      `json:"size"`
	StoragePath string     `gorm:"size:255" json:"storage_path"`
	Error       string     `gorm:"type:text" json:"error"`
}

type RestoreTask struct {
	BaseModel
	BackupTask   BackupTask `gorm:"foreignKey:BackupTaskID" json:"backup_task"`
	BackupTaskID uint       `json:"backup_task_id"`
	Status       string     `gorm:"size:20;not null" json:"status"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Error        string     `gorm:"type:text" json:"error"`
}

func (BackupTask) TableName() string {
	return "backup_tasks"
}

func (RestoreTask) TableName() string {
	return "restore_tasks"
}
