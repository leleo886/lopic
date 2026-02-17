package cli

import (
	"fmt"
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/database"
	"github.com/leleo886/lopic/services/admin_services"
)

// ResetAdminPassword 重置管理员密码
func ResetAdminPassword(newPassword string) error {
	// 加载配置
	appConfig, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}


	// 连接数据库
	db, err := database.Connect(&appConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 创建用户服务实例
	userService := admin_services.NewUserService(db, appConfig)

	// 查找role为admin的用户
	type User struct {
		ID       int
		Username string
		RoleName string
		Email    string
		Active   bool
	}
	var user User
	result := db.Table("users").Select("users.id, users.username, roles.name as role_name, users.email, users.active").Joins("JOIN roles ON users.role_id = roles.id").Where("roles.name = ?", "admin").First(&user)
	if result.Error != nil {
		return fmt.Errorf("admin user not found: %w", result.Error)
	}

	// 构建用户请求
	req := admin_services.UserRequest{
		Username: user.Username, // 保持用户名不变
		Password: newPassword,   // 更新密码
		Role:     user.RoleName, // 保持角色不变
		Email:    user.Email,    // 保持邮箱不变
		Active:   user.Active,   // 保持激活状态不变
	}

	// 调用UpdateUser方法更新密码
	updatedUser, err := userService.UpdateUser(user.ID, 0, req)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	fmt.Printf("Success: Password changed for admin user '%s'\n", updatedUser.Username)
	return nil
}
