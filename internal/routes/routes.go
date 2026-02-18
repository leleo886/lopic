package routes

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/controllers"
	"github.com/leleo886/lopic/controllers/admin_controllers"
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/mail"
	"github.com/leleo886/lopic/internal/websocket"
	"github.com/leleo886/lopic/middleware"
	"github.com/leleo886/lopic/services"
	"github.com/leleo886/lopic/services/admin_services"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:embed dist/*
var frontendFS embed.FS

// SetupRouter 设置路由
func SetupRouter(config *config.Config,
	hub *websocket.Hub,
	imageService *services.ImageService,
	mailService *mail.MailService,
	authService *services.AuthService,
	albumService *services.AlbumService,
	userService *services.UserService,
	backupService *admin_services.BackupService,
	adminRoleService *admin_services.RoleService,
	adminUserService *admin_services.UserService,
	adminImageService *admin_services.ImageService,
	adminAlbumService *admin_services.AlbumService,
	adminStorageService *admin_services.StorageService,
	galleryService *services.GalleryService,
) *gin.Engine {
	gin.SetMode(config.Server.Mode)

	// 创建GIN引擎
	router := gin.Default()

	// 配置CORS
	oconfig := cors.DefaultConfig()
	if len(config.Server.AllowOrigins) > 0 {
		oconfig.AllowOrigins = config.Server.AllowOrigins
		oconfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		oconfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie", "X-CSRF-TOKEN"}
		oconfig.ExposeHeaders = []string{"Content-Length"}

		// 安全检查：允许凭证时必须指定具体的来源，不能使用 *
		hasWildcard := false
		for _, origin := range config.Server.AllowOrigins {
			if origin == "*" {
				hasWildcard = true
				break
			}
		}
		if hasWildcard {
			fmt.Println("Warning: CORS configuration contains '*', which is insecure when AllowCredentials is true. Removing credentials support.")
			log.Warn("CORS configuration contains '*', which is insecure when AllowCredentials is true. Removing credentials support.")
			oconfig.AllowCredentials = false
		} else {
			oconfig.AllowCredentials = true
		}

		router.Use(cors.New(oconfig))
	}

	// 创建处理器
	imageController := controllers.NewImageController(imageService, hub)
	albumController := controllers.NewAlbumController(albumService)
	authController := controllers.NewAuthController(authService, config)
	userController := controllers.NewUserController(userService)
	adminRoleController := admin_controllers.NewRoleController(adminRoleService)
	adminUserController := admin_controllers.NewUserController(userService, adminUserService, hub)
	adminImageController := admin_controllers.NewImageController(adminImageService, hub)
	adminAlbumController := admin_controllers.NewAlbumController(adminAlbumService)
	adminSystemController := admin_controllers.NewSystemController(mailService, &config.SystemSettings.General, &config.SystemSettings.Gallery)
	galleryController := controllers.NewGalleryController(galleryService, &config.SystemSettings.Gallery)
	backupController := admin_controllers.NewBackupController(backupService)
	adminStorageController := admin_controllers.NewStorageController(adminStorageService)

	// 配置Swagger
	if config.Swagger.Enabled {
		router.GET(config.Swagger.Path+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 健康检查路由
	if config.Server.Mode == gin.DebugMode {
		router.GET("/health", controllers.HealthCheck)
	}

	// 上传进度中间件
	uploadProgressMiddleware := middleware.NewUploadProgressMiddleware(hub)

	// 前端页面 (从嵌入的文件系统提供)
	assetsFS, _ := fs.Sub(frontendFS, "dist/assets")
	router.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.FileFromFS(filepath, http.FS(assetsFS))
	})

	indexHTML, _ := frontendFS.ReadFile("dist/index.html")
	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	router.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	// 静态文件服务
	router.Static(config.Server.StaticPath, config.Server.UploadDir)

	galleryGroup := router.Group("/api/gallery")
	galleryGroup.GET("/config", galleryController.GetGalleryConfig)
	galleryGroup.Use(middleware.GalleryPermission())
	galleryGroup.Use(middleware.RateLimit(&middleware.RateLimitConfig{
		Limit:      80,          // 1分钟80次请求
		WindowSize: time.Minute, // 1分钟窗口
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP() // 使用客户端IP作为限制键
		},
	}))
	{
		galleryGroup.GET("/albums/:user_name", galleryController.GetGallerys)
		galleryGroup.GET("/images/:user_name/:album_id", galleryController.GetGalleryImages)
		galleryGroup.GET("/search/:user_name", galleryController.SearchGalleryImages)
	}

	// 认证路由组
	authGroup := router.Group("/api/auth")
	authGroup.Use(middleware.RateLimit(&middleware.RateLimitConfig{
		Limit:      5,           // 5次请求
		WindowSize: time.Minute, // 1分钟窗口
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP() // 使用客户端IP作为限制键
		},
	}))
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/reset-password/request", authController.RequestPasswordReset)
		authGroup.POST("/reset-password", authController.ResetPassword)
		authGroup.GET("/verify-email", authController.VerifyEmail)
		authGroup.POST("/refresh", authController.RefreshToken)
		authGroup.POST("/logout", authController.Logout)
	}

	// 需要认证的API路由组
	// WebSocket路由
	wsGroup := router.Group("/ws")
	wsGroup.Use(middleware.JWT(&config.JWT))
	{
		wsGroup.GET("/upload", websocket.HandleWebSocket(hub))
	}

	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.JWT(&config.JWT))
	apiGroup.Use(middleware.Casbin())
	apiGroup.Use(middleware.RateLimit(&middleware.RateLimitConfig{
		Limit:      80,          // 1分钟80次请求
		WindowSize: time.Minute, // 1分钟窗口
		KeyFunc: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				return fmt.Sprintf("user:%v", userID) // 使用用户ID作为限制键
			}
			return c.ClientIP() // 后备使用客户端IP
		},
	}))
	{
		// 用户路由
		userGroup := apiGroup.Group("/users")
		{
			userGroup.GET("/me", userController.GetMe)
			userGroup.PUT("/me", userController.UpdateMe)
			userGroup.GET("/me/storage", userController.GetStorageUsage)
			userGroup.GET("/me/tags-cloud", userController.GetImagesTagsCloud)
		}

		// 图片路由
		imageGroup := apiGroup.Group("/images")
		{
			imageGroup.POST("/upload", uploadProgressMiddleware.Handle(), imageController.UploadImage)
			imageGroup.GET("", imageController.GetImages)
			imageGroup.GET("/:id", imageController.GetImage)
			imageGroup.PUT("", imageController.UpdateImage)
			imageGroup.DELETE("", imageController.DeleteImage)
			imageGroup.POST("/albums", imageController.AddImageToAlbum)
			imageGroup.DELETE("/albums", imageController.RemoveImageFromAlbum)
			imageGroup.GET("/search", imageController.SearchImagesByTagsOrTitle)
		}

		// 相册路由
		albumGroup := apiGroup.Group("/albums")
		{
			albumGroup.POST("", albumController.CreateAlbum)
			albumGroup.GET("", albumController.GetAlbums)
			albumGroup.GET("/:id", albumController.GetAlbum)
			albumGroup.PUT("/:id", albumController.UpdateAlbum)
			albumGroup.DELETE("/:id", albumController.DeleteAlbum)
			albumGroup.GET("/:id/images", albumController.GetAlbumImages)
			albumGroup.GET("/images/not-in-any", albumController.GetNotInAnyAlbum)
		}

		// 管理员路由
		adminGroup := apiGroup.Group("/admin")
		{
			adminUserGroup := adminGroup.Group("/users")
			{
				adminUserGroup.GET("", adminUserController.GetUsers)
				adminUserGroup.POST("", adminUserController.CreateUser)
				adminUserGroup.GET("/:id", adminUserController.GetUser)
				adminUserGroup.PUT("/:id", adminUserController.UpdateUser)
				adminUserGroup.DELETE("/:id", adminUserController.DeleteUser)
				adminUserGroup.GET("/tags-cloud", adminUserController.GetAllImagesTagsCloud)
				adminUserGroup.GET("/:id/tags-cloud", adminUserController.GetUserImagesTagsCloud)
			}

			adminRoleGroup := adminGroup.Group("/roles")
			{
				adminRoleGroup.GET("", adminRoleController.GetRoles)
				adminRoleGroup.GET("/:id", adminRoleController.GetRole)
				adminRoleGroup.POST("", adminRoleController.CreateRole)
				adminRoleGroup.PUT("/:id", adminRoleController.UpdateRole)
				adminRoleGroup.DELETE("/:id", adminRoleController.DeleteRole)
				adminRoleGroup.GET("/users-count", adminRoleController.GetUsersCountByRole)
			}

			adminImageGroup := adminGroup.Group("/images")
			{
				adminImageGroup.GET("", adminImageController.GetAllImages)
				adminImageGroup.GET("/:id", adminImageController.GetImage)
				adminImageGroup.DELETE("", adminImageController.DeleteImage)
				adminImageGroup.PUT("/storagename", adminImageController.UpdateImageStorage)
			}

			adminAlbumGroup := adminGroup.Group("/albums")
			{
				adminAlbumGroup.GET("", adminAlbumController.GetAllAlbums)
				adminAlbumGroup.GET("/:id", adminAlbumController.GetAlbum)
				adminAlbumGroup.DELETE("", adminAlbumController.DeleteAlbum)
			}

			adminSystemGroup := adminGroup.Group("/system")
			{
				adminSystemGroup.GET("/info", adminSystemController.GetSystemInfo)
				adminSystemGroup.PUT("/info", adminSystemController.UpdateSystemInfo)
			}

			adminBackupGroup := adminGroup.Group("/backup")
			{
				adminBackupGroup.POST("", backupController.CreateBackup)
				adminBackupGroup.GET("/list", backupController.GetBackupList)
				adminBackupGroup.GET("/restore/list", backupController.GetRestoreRecords)
				adminBackupGroup.DELETE("/:id", backupController.DeleteBackup)
				adminBackupGroup.POST("/restore/:id", backupController.RestoreBackup)
				adminBackupGroup.DELETE("/restore/:id", backupController.DeleteRestoreTask)
				adminBackupGroup.GET("/download/:id", backupController.DownloadBackup)
				adminBackupGroup.POST("/upload", uploadProgressMiddleware.Handle(), backupController.UploadBackup)
			}

			adminStorageGroup := adminGroup.Group("/storages")
			{
				adminStorageGroup.GET("", adminStorageController.GetStorages)
				adminStorageGroup.GET("/:id", adminStorageController.GetStorage)
				adminStorageGroup.GET("/name/:name", adminStorageController.GetStorageByName)
				adminStorageGroup.POST("", adminStorageController.CreateStorage)
				adminStorageGroup.PUT("/:id", adminStorageController.UpdateStorage)
				adminStorageGroup.DELETE("/:id", adminStorageController.DeleteStorage)
				adminStorageGroup.POST("/test", adminStorageController.TestStorageConnection)
			}
		}
	}

	return router
}
