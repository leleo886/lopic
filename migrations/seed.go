package migrations

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB, serverConfig *config.ServerConfig) error {
	log.Info("Starting to seed database...")

	if err := InitializeRoles(db); err != nil {
		return err
	}

	if err := InitializeAdminUser(db); err != nil {
		return err
	}

	if err := InitializeSystemSettings(db); err != nil {
		return err
	}

	if err := InitializeLocalStorage(db, serverConfig); err != nil {
		return err
	}

	log.Info("Database seeding completed successfully")
	return nil
}

func InitializeRoles(db *gorm.DB) error {
	log.Info("Initializing default roles...")

	roles := []models.Role{
		{
			Name:              "admin",
			Description:       "Administrator role",
			AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff", ".tif"},
			MaxFilesPerUpload: -1,
			MaxFileSizeMB:     -1,
			MaxAlbumsPerUser:  -1,
			MaxStorageSizeMB:  -1,
			GalleryOpen:       true,
		},
		{
			Name:              "user",
			Description:       "Regular user role",
			AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
			MaxFilesPerUpload: 10,
			MaxFileSizeMB:     5,
			MaxAlbumsPerUser:  6,
			MaxStorageSizeMB:  300,
			GalleryOpen:       false,
		},
	}

	for _, role := range roles {
		var existingRole models.Role
		result := db.Where("name = ?", role.Name).First(&existingRole)
		if result.RowsAffected == 0 {
			log.Infof("Initializing role: %s", role.Name)
			result = db.Create(&role)
			if result.Error != nil {
				log.Errorf("Initialization failed: %s, Error: %v", role.Name, result.Error)
				return result.Error
			} else {
				log.Infof("Initialization successful: %s (ID: %d)", role.Name, role.ID)
			}
		} else {
			log.Infof("Role already exists: %s (ID: %d)", role.Name, existingRole.ID)
		}
	}

	return nil
}

func InitializeAdminUser(db *gorm.DB) error {
	log.Info("Initializing admin user...")

	var adminUser models.User
	result := db.Joins("Role").Where("Role.name = ?", "admin").First(&adminUser)
	if result.Error == nil {
		log.Infof("Admin user already exists (ID: %d)", adminUser.ID)
		return nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		log.Errorf("Failed to check admin user: %v", result.Error)
		return result.Error
	}

	var adminRole models.Role
	result = db.Where("name = ?", "admin").First(&adminRole)
	if result.RowsAffected == 0 {
		log.Errorf("Admin role does not exist, please initialize roles first")
		return gorm.ErrRecordNotFound
	}

	initialPWD := generateRandomPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(initialPWD), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("Failed to hash password: %v", err)
		return err
	}

	adminUser = models.User{
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@example.com",
		RoleID:   adminRole.ID,
		Active:   true,
	}

	result = db.Create(&adminUser)
	if result.Error != nil {
		log.Errorf("Failed to create admin user: %v", result.Error)
		return result.Error
	}

	log.Infof("Admin user created successfully (ID: %d)", adminUser.ID)

	fmt.Printf("Admin user initialized successfully\n")
	fmt.Printf("Username: admin\n")
	fmt.Printf("Password: %s\n", initialPWD)

	return nil
}

func InitializeSystemSettings(db *gorm.DB) error {
	log.Info("Initializing system settings...")

	// 检查是否已存在系统设置
	var existingSetting models.SystemSetting
	result := db.Model(&models.SystemSetting{}).First(&existingSetting)
	if result.RowsAffected > 0 {
		log.Info("System settings already exist")
		return nil
	}

	settings := models.SystemSettings{
		General: models.GeneralConfig{
			MaxThumbSize:    800,
			RegisterEnabled: false,
			MaxTags:         60,
		},
		Mail: models.MailConfig{
			Enabled:       false,
			ServerAddress: "",
			SMTP: models.SMTPConfig{
				Host:     "",
				Port:     587,
				Username: "",
				Password: "",
				From:     "",
				FromName: "",
			},
		},
		Gallery: models.GalleryConfig{
			Title:           "LOPIC",
			BackgroundImage: "",
		},
	}

	systemSetting := models.SystemSetting{
		Value: settings,
	}

	// 创建新记录
	if err := db.Create(&systemSetting).Error; err != nil {
		return err
	}

	log.Info("System settings initialized successfully")
	return nil
}

func InitializeLocalStorage(db *gorm.DB, serverConfig *config.ServerConfig) error {
	// 检查 local 存储是否已存在
	var existingStorage models.Storage
	result := db.Where("name = ?", "local").First(&existingStorage)
	if result.RowsAffected > 0 {
		// 已存在，无需创建
		return nil
	}

	// 创建默认的 local 存储
	config := models.StorageConfig{
		BasePath:  serverConfig.UploadDir,
		StaticURL: serverConfig.StaticPath,
	}

	storage := &models.Storage{
		Name:   "local",
		Type:   "local",
		Config: config,
	}

	result = db.Create(storage)
	if result.Error != nil {
		log.Errorf("failed to create local storage: error=%v", result.Error)
		return result.Error
	}
	return nil
}

func generateRandomPassword() string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	password := make([]byte, 8)
	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		password[i] = charset[n.Int64()]
	}
	return string(password)
}
