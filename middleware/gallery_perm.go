package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/internal/database"
	"github.com/leleo886/lopic/internal/log"
	cerrors "github.com/leleo886/lopic/internal/error"
	"gorm.io/gorm"
)

func GalleryPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var result *gorm.DB
		userName := c.Param("user_name")
		
		if userName == "" {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}else if userName == "$admin$"{
			roleName := "admin"
			result = database.GetDB().
				Preload("Role").
				Joins("LEFT JOIN roles ON users.role_id = roles.id").
				Where("roles.name = ?", roleName).
				First(&user)

			if result.RowsAffected <= 0{
				statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUserNotFound)
				c.JSON(statusCode, errorResponse)
				c.Abort()
				return
			}
		}else {
			// 获取用户角色
			result = database.GetDB().Preload("Role").First(&user, "username = ?", userName)
			// 如果用户不存在
			if result.RowsAffected <= 0{
				statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUserNotFound)
				c.JSON(statusCode, errorResponse)
				c.Abort()
				return
			}
		}

		if result.Error != nil {
			log.Errorf("failed to get user: username=%s, error=%v", userName, result.Error)
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrInternalServer)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		// 检查用户角色是否有画廊权限
		if !user.Role.GalleryOpen {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrGalleryPermissionDenied)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Next()
	}
}