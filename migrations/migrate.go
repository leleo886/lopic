package migrations

import (
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	log.Infof("Starting database migration...")

	if err := db.AutoMigrate(&models.Role{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Album{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Image{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.ImageAlbum{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.PasswordResetCode{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.SystemSetting{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.RefreshTokenBlacklist{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.BackupTask{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.RestoreTask{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Storage{}); err != nil {
		return err
	}

	log.Infof("Database migration completed successfully")
	return nil
}
