package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/database"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
)

// JWT自定义声明结构体
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleName string `json:"role_name"`
	RoleID   uint   `json:"role_id"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims 刷新令牌声明结构体
type RefreshTokenClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	RoleID    uint   `json:"role_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username string, roleID uint, jwtConfig *config.JWTConfig) (string, error) {
	// 获取角色名称
	var role models.Role
	result := database.GetDB().First(&role, roleID)
	if result.Error != nil {
		log.Errorf("role not found: role_id=%d", roleID)
		return "", cerrors.ErrRoleNotFound
	}

	// 创建声明
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RoleName: role.Name,
		RoleID:   role.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtConfig.Expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    jwtConfig.Issuer,
			Subject:   username,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的编码后的字符串令牌
	tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		log.Errorf("error generating token: %v", err)
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string, jwtConfig *config.JWTConfig) (*JWTClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, cerrors.ErrInvalidSigningMethod
		}
		return []byte(jwtConfig.Secret), nil
	})

	if err != nil {
		log.Errorf("error parsing token: %v", err)
		return nil, err
	}

	// 验证令牌有效性
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, cerrors.ErrInvalidToken
}

// ValidateToken 验证JWT令牌是否有效
func ValidateToken(tokenString string, jwtConfig *config.JWTConfig) bool {
	_, err := ParseToken(tokenString, jwtConfig)
	return err == nil
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID uint, username string, roleID uint, jwtConfig *config.JWTConfig) (string, error) {
	// 创建刷新令牌声明
	claims := RefreshTokenClaims{
		UserID:    userID,
		Username:  username,
		RoleID:    roleID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtConfig.RefreshTokenExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    jwtConfig.Issuer,
			Subject:   username,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整的编码后的字符串令牌
	tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		log.Errorf("error generating refresh token: %v", err)
		return "", err
	}

	return tokenString, nil
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string, jwtConfig *config.JWTConfig) (*RefreshTokenClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, cerrors.ErrInvalidSigningMethod
		}
		return []byte(jwtConfig.Secret), nil
	})

	if err != nil {
		log.Errorf("error parsing refresh token: %v", err)
		return nil, err
	}

	// 验证令牌有效性
	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		// 验证令牌类型
		if claims.TokenType != "refresh" {
			log.Errorf("invalid token type: expected 'refresh', got '%s'", claims.TokenType)
			return nil, cerrors.ErrInvalidToken
		}
		return claims, nil
	}

	return nil, cerrors.ErrInvalidToken
}

// ValidateRefreshToken 验证刷新令牌是否有效
func ValidateRefreshToken(tokenString string, jwtConfig *config.JWTConfig) bool {
	_, err := ParseRefreshToken(tokenString, jwtConfig)
	return err == nil
}

// RefreshAccessToken 使用刷新令牌生成新的访问令牌
func RefreshAccessToken(refreshTokenString string, jwtConfig *config.JWTConfig) (string, error) {
	// 解析刷新令牌
	claims, err := ParseRefreshToken(refreshTokenString, jwtConfig)
	if err != nil {
		return "", err
	}

	// 生成新的访问令牌
	return GenerateToken(claims.UserID, claims.Username, claims.RoleID, jwtConfig)
}
