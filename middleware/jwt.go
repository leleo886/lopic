package middleware

import (
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/utils"
	"github.com/leleo886/lopic/internal/database"
	"github.com/leleo886/lopic/models"
)

// JWT 认证中间件
func JWT(config *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 如果没有Authorization头，那就再查?token
			token := c.Query("token")
			if token == "" {
				statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
				c.JSON(statusCode, errorResponse)
				c.Abort()
				return
			}
			authHeader = token
		}

		// 检查Authorization格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrForbidden)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		// 解析JWT令牌
		claims, err := utils.ParseToken(parts[1], config)
		if err != nil {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrInvalidToken)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}
		
		var user models.User
		res := database.GetDB().Preload("Role").First(&user, claims.UserID)
		if res.RowsAffected == 0 {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}
		
		if user.Role.ID != claims.RoleID && user.Role.Name != claims.RoleName {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		if !user.Active {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUserNotActive)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.RoleName)

		c.Next()
	}
}
