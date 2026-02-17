package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/internal/mail"
	"github.com/leleo886/lopic/models"
	"github.com/leleo886/lopic/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db          *gorm.DB
	mailService *mail.MailService
	cfg         *config.Config
}

type LoginResponse struct {
	User          models.User   `json:"user"`
	TokenResponse TokenResponse `json:"token_response"`
}

// TokenResponse 令牌响应结构体
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

func NewAuthService(db *gorm.DB, mailService *mail.MailService, cfg *config.Config) *AuthService {
	return &AuthService{db: db, mailService: mailService, cfg: cfg}
}

func (s *AuthService) Register(username, password, email, locale string) (string, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	result := s.db.Where("username = ?", username).First(&existingUser)
	if result.RowsAffected > 0 {
		return "", cerrors.ErrUserAlreadyExists
	}

	// 检查邮箱是否已存在
	result = s.db.Where("email = ?", email).First(&existingUser)
	if result.RowsAffected > 0 {
		return "", cerrors.ErrUserAlreadyExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", cerrors.ErrPwdEncFailed
	}

	// 获取user角色的ID
	var userRole models.Role
	result = s.db.Where("name = ?", "user").First(&userRole)
	if result.RowsAffected == 0 {
		return "", cerrors.ErrUserRoleNotFound
	}

	// 创建用户
	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		RoleID:   userRole.ID,
		Active:   !s.mailService.IsEnabled(),
	}

	result = s.db.Create(&user)
	if result.Error != nil {
		return "", cerrors.ErrCreateUserFailed
	}

	if s.mailService.IsEnabled() {
		verifyToken := s.generateVerifyToken(email)
		verifyLink := fmt.Sprintf("%s/api/auth/verify-email?token=%s", s.mailService.GetServerAddress(), verifyToken)
		go func() {
			if err := s.mailService.SendEmailVerification(email, username, verifyLink, locale); err != nil {
				log.Errorf("failed to send email verification: email=%s, error=%v", email, err)
			}
		}()
		return "Register_EmailSent", nil
	}

	return "Register_NoMail", nil
}

func (s *AuthService) Login(username, password string, jwtCfg *config.JWTConfig) (*LoginResponse, error) {
	// 查找用户
	var user models.User
	result := s.db.Preload("Role").Where("username = ?", username).First(&user)
	if result.RowsAffected == 0 {
		return nil, cerrors.ErrInvalidCredentials
	}

	// 验证active
	if !user.Active {
		return nil, cerrors.ErrInvalidCredentials
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, cerrors.ErrInvalidCredentials
	}

	// 生成JWT令牌
	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.RoleID, jwtCfg)
	if err != nil {
		return nil, cerrors.ErrGenerateTokenFailed
	}

	// 生成刷新令牌
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username, user.RoleID, jwtCfg)
	if err != nil {
		return nil, cerrors.ErrGenerateTokenFailed
	}

	return &LoginResponse{
		User: user,
		TokenResponse: TokenResponse{
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			ExpiresIn:        jwtCfg.Expire,
			RefreshExpiresIn: jwtCfg.RefreshTokenExpire,
		},
	}, nil
}

func (s *AuthService) RequestPasswordReset(email, locale string) error {
	if !s.mailService.IsEnabled() {
		return cerrors.ErrMailServiceDisabled
	}

	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.RowsAffected == 0 {
		return nil
	}

	// 生成6位数字验证码
	code := generateVerificationCode()

	// 删除该邮箱之前的所有验证码
	s.db.Where("email = ?", email).Delete(&models.PasswordResetCode{})

	// 保存新验证码到数据库
	resetCode := models.PasswordResetCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	// 保存新验证码到数据库
	result = s.db.Create(&resetCode)
	if result.Error != nil {
		return cerrors.ErrCreateUserFailed
	}

	// 异步发送验证码邮件，不阻塞主流程
	go func() {
		if err := s.mailService.SendResetPasswordCode(email, code, locale); err != nil {
			log.Errorf("failed to send reset password code: email=%s, error=%v", email, err)
		}
	}()

	return nil
}

func (s *AuthService) VerifyPasswordResetCode(email, code string) error {
	var resetCode models.PasswordResetCode
	result := s.db.Where("email = ? AND code = ? AND used = ? AND expires_at > ?", email, code, false, time.Now()).First(&resetCode)
	if result.RowsAffected == 0 {
		return cerrors.ErrResetTokenInvalid
	}

	// 标记验证码为已使用
	resetCode.Used = true
	result = s.db.Save(&resetCode)
	if result.Error != nil {
		return cerrors.ErrCreateUserFailed
	}
	return nil
}

func (s *AuthService) ResetPassword(email, code, newPassword string) error {
	if !s.mailService.IsEnabled() {
		return cerrors.ErrMailServiceDisabled
	}

	// 验证邮箱是否存在
	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.RowsAffected == 0 {
		return cerrors.ErrInvalidResetCode
	}

	// 验证验证码
	err := s.VerifyPasswordResetCode(email, code)
	if err != nil {
		log.Errorf("failed to verify reset password code: email=%s, error=%v", email, err)
		return cerrors.ErrInvalidResetCode
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("failed to encrypt new password: email=%s, error=%v", email, err)
		return cerrors.ErrPwdEncFailed
	}

	// 更新用户密码
	result = result.Update("password", string(hashedPassword))
	if result.Error != nil {
		log.Errorf("failed to update user password: email=%s, error=%v", email, result.Error)
		return cerrors.ErrCreateUserFailed
	}

	return nil
}

func generateVerificationCode() string {
	code := make([]byte, 3)
	if _, err := rand.Read(code); err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%03x", code)
}

func (s *AuthService) VerifyEmail(token string) error {
	if !s.mailService.IsEnabled() {
		return cerrors.ErrMailServiceDisabled
	}

	email, err := s.validateVerifyToken(token)
	if err != nil {
		return err
	}

	user := models.User{}
	result := s.db.Model(&user).Where("email = ?", email).First(&user)
	if result.Error != nil {
		return cerrors.ErrCreateUserFailed
	}

	if user.Active {
		return cerrors.ErrUserAlreadyActive
	}

	result = result.Update("active", true)
	if result.Error != nil {
		return cerrors.ErrCreateUserFailed
	}

	return nil
}

func (s *AuthService) generateVerifyToken(email string) string {
	return utils.GenerateSignedToken(email, time.Now().Add(24*time.Hour).Unix(), "verify", &s.cfg.JWT)
}

func (s *AuthService) validateVerifyToken(token string) (string, error) {
	email, timestamp, err := utils.ValidateSignedToken(token, "verify", &s.cfg.JWT)
	if err != nil {
		return "", cerrors.ErrResetTokenInvalid
	}

	if utils.IsTokenExpired(timestamp, 24) {
		return "", cerrors.ErrResetTokenExpired
	}

	return email, nil
}

func (s *AuthService) RefreshToken(refreshToken string, jwtCfg *config.JWTConfig) (*TokenResponse, error) {
	// 检查令牌是否在黑名单中
	inBlacklist, err := IsInBlacklist(s.db, refreshToken)
	if err != nil {
		return nil, cerrors.ErrInternalServer
	}
	if inBlacklist {
		return nil, cerrors.ErrInvalidToken
	}

	// 验证刷新令牌并生成新的访问令牌
	newAccessToken, err := utils.RefreshAccessToken(refreshToken, jwtCfg)
	if err != nil {
		return nil, cerrors.ErrInvalidToken
	}

	// 解析刷新令牌以获取用户信息
	refreshTokenClaims, err := utils.ParseRefreshToken(refreshToken, jwtCfg)
	if err != nil {
		return nil, cerrors.ErrInvalidToken
	}

	// 生成新的刷新令牌
	newRefreshToken, err := utils.GenerateRefreshToken(
		refreshTokenClaims.UserID,
		refreshTokenClaims.Username,
		refreshTokenClaims.RoleID,
		jwtCfg,
	)
	if err != nil {
		return nil, cerrors.ErrGenerateTokenFailed
	}

	// 将旧的刷新令牌添加到黑名单
	if err := AddToBlacklist(s.db, refreshToken, refreshTokenClaims.ExpiresAt.Time); err != nil {
		// 记录错误但不影响主要流程
		log.Errorf("failed to add token to blacklist: %v", err)
	}

	return &TokenResponse{
		AccessToken:      newAccessToken,
		RefreshToken:     newRefreshToken,
		ExpiresIn:        jwtCfg.Expire,
		RefreshExpiresIn: jwtCfg.RefreshTokenExpire,
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(refreshToken string, jwtCfg *config.JWTConfig) error {
	// 检查令牌是否在黑名单中
	inBlacklist, err := IsInBlacklist(s.db, refreshToken)
	if err != nil {
		return cerrors.ErrInternalServer
	}
	if inBlacklist {
		return cerrors.ErrInvalidToken
	}

	// 解析刷新令牌以获取过期时间
	refreshTokenClaims, err := utils.ParseRefreshToken(refreshToken, jwtCfg)
	if err != nil {
		return cerrors.ErrInvalidToken
	}

	// 将刷新令牌添加到黑名单
	if err := AddToBlacklist(s.db, refreshToken, refreshTokenClaims.ExpiresAt.Time); err != nil {
		log.Errorf("failed to add token to blacklist during logout: %v", err)
		return cerrors.ErrInternalServer
	}

	return nil
}

// AddToBlacklist
func AddToBlacklist(db *gorm.DB, token string, expiration time.Time) error {
	// 计算令牌的哈希值，减少存储大小
	tokenHash := calculateTokenHash(token)
	blacklistEntry := models.RefreshTokenBlacklist{
		TokenHash: tokenHash,
		ExpiresAt: expiration,
	}
	return db.Create(&blacklistEntry).Error
}

// IsInBlacklist
func IsInBlacklist(db *gorm.DB, token string) (bool, error) {
	tokenHash := calculateTokenHash(token)
	var count int64
	err := db.Model(&models.RefreshTokenBlacklist{}).Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).Count(&count).Error
	return count > 0, err
}

// CleanupBlacklist
func CleanupBlacklist(db *gorm.DB) error {
	return db.Where("expires_at <= ?", time.Now()).Delete(&models.RefreshTokenBlacklist{}).Error
}

// calculateTokenHash
func calculateTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
