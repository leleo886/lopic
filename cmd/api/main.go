package main

import (
	"flag"
	"fmt"
	"time"
	"github.com/leleo886/lopic/cmd/api/cli"
	_ "github.com/leleo886/lopic/docs"
	"github.com/leleo886/lopic/internal/casbin"
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/database"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/mail"
	"github.com/leleo886/lopic/internal/routes"
	"github.com/leleo886/lopic/internal/storage"
	"github.com/leleo886/lopic/internal/websocket"
	"github.com/leleo886/lopic/migrations"
	"github.com/leleo886/lopic/services"
	"github.com/leleo886/lopic/services/admin_services"
)

// @title Lopic API
// @version 1.0.0
// @description RESTful API 

// @host localhost:6060
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	var (
		resetPwd    string
		startServer bool
		configPath  string
	)

	flag.StringVar(&resetPwd, "resetpwd", "", "Reset admin password: --resetpwd=<new-password>")
	flag.BoolVar(&startServer, "serve", false, "Start server: --serve")
	flag.StringVar(&configPath, "config", "configs/config.yaml", "Configuration file path: --config=<path>")
	flag.Parse()

	if resetPwd != "" {
		if err := cli.ResetAdminPassword(resetPwd); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		return
	}

	if !startServer {
		fmt.Println("Start server: --serve")
		fmt.Println("Reset admin password: --resetpwd=<new-password>")
		return
	}

	// 加载配置
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 初始化日志
	log.InitLogger(
		appConfig.Log.Level,
		appConfig.Log.OutputPath,
		appConfig.Log.MaxSize,
		appConfig.Log.MaxBackups,
		appConfig.Log.MaxAge,
		appConfig.Log.Compress,
		appConfig.Log.ConsoleOutput,
	)

	// 连接数据库
	db, err := database.Connect(&appConfig.Database)
	if err != nil {
		fmt.Println("Failed to connect to database. Please check your configuration.")
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := migrations.Migrate(db); err != nil {
		fmt.Println("Database migration failed. Please check your configuration.")
		log.Fatalf("Database migration failed: %v", err)
		
	}

	if err := migrations.Seed(db, &appConfig.Server); err != nil {
		fmt.Println("Seed data initialization failed. Please check your configuration.")
		log.Fatalf("Seed data initialization failed: %v", err)
		
	}

	// 清理过期的刷新令牌黑名单条目
	if err := services.CleanupBlacklist(db); err != nil {
		log.Errorf("Failed to cleanup refresh token blacklist: %v", err)
	} else {
		log.Infof("Successfully cleaned up refresh token blacklist")
	}

	// 初始化WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	_, err = casbin.InitCasbin()
	if err != nil {
		log.Fatalf("Failed to initialize Casbin: %v", err)
	}

	// 加载系统设置
	systemSettings, err := config.LoadSystemSettingsFromDatabase(db)
	if err != nil {
		fmt.Println("Failed to load system settings. Please check your configuration.")
		log.Fatalf("Failed to load system settings: %v", err)
		
	}

	appConfig.SystemSettings = systemSettings

	// 创建服务
	storageService := storage.NewStorageService(appConfig)
	imageService := services.NewImageService(db, appConfig)
	mailService := mail.NewMailService(&systemSettings.Mail)
	authService := services.NewAuthService(db, mailService, appConfig)
	albumService := services.NewAlbumService(db)
	userService := services.NewUserService(db, appConfig)
	backupService := admin_services.NewBackupService(db, storageService, &appConfig.Database, &appConfig.Server)
	adminRoleService := admin_services.NewRoleService(db)
	adminUserService := admin_services.NewUserService(db, appConfig)
	adminImageService := admin_services.NewImageService(db, appConfig)
	adminAlbumService := admin_services.NewAlbumService(db)
	galleryService := services.NewGalleryService(db)
	adminStorageService := admin_services.NewStorageService(db)

	// 设置路由
	router := routes.SetupRouter(appConfig, hub,
		imageService, mailService, authService, albumService, userService, backupService,
		adminRoleService, adminUserService, adminImageService, adminAlbumService, adminStorageService, galleryService)

	// 启动定期清理过期刷新令牌黑名单的goroutine
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // 每24小时执行一次
		defer ticker.Stop()
		for range ticker.C {
			log.Infof("Running periodic cleanup of refresh token blacklist")
			if err := services.CleanupBlacklist(db); err != nil {
				log.Errorf("Failed to cleanup refresh token blacklist: %v", err)
			} else {
				log.Infof("Successfully cleaned up refresh token blacklist")
			}
		}
	}()

	// 启动服务器
	serverAddr := fmt.Sprintf(":%d", appConfig.Server.Port)
	fmt.Printf("Server started at http://localhost%s\n", serverAddr)

	if appConfig.Swagger.Enabled {
		fmt.Printf("Swagger documentation available at http://localhost%s%s/index.html\n", serverAddr, appConfig.Swagger.Path)
	}

	if err := router.Run(serverAddr); err != nil {
		fmt.Println("Server failed to start. Please check your configuration.")
		log.Fatalf("Server failed to start: %v", err)
		
	}
}
