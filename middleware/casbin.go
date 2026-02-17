package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/internal/casbin"
	"github.com/leleo886/lopic/internal/database"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/models"
)

const AdminRoleName = "admin"
const UserRoleName = "user"

func Casbin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		var user models.User
		db := database.GetDB()
		if err := db.Preload("Role").First(&user, userID).Error; err != nil {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUserNotFound)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		roleName := user.Role.Name

		if roleName != AdminRoleName {
			roleName = UserRoleName
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		allowed, err := casbin.GetEnforcer().Enforce(roleName, path, method)
		if err != nil {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrInternalServer)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		if !allowed {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrForbidden)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}

		c.Next()
	}
}
