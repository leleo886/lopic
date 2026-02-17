package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/internal/config"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/success"
	"github.com/leleo886/lopic/services"
)

const (
	successHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Success</title>
</head>
<body>
    <h1 style="color: green;">%s</h1>
</body>
</html>`

	errorHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error</title>
</head>
<body>
    <h1>%s</h1>
</body>
</html>`
)

type AuthController struct {
	authService *services.AuthService
	cfg         *config.Config
}

func NewAuthController(authService *services.AuthService, cfg *config.Config) *AuthController {
	return &AuthController{authService: authService, cfg: cfg}
}

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Locale   string `json:"locale" binding:"required,oneof=zh en"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LogoutRequest 登出请求结构体
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RequestPasswordResetRequest 请求密码重置结构体
type RequestPasswordResetRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Locale string `json:"locale" binding:"required,oneof=zh en"`
}

// ResetPasswordRequest 重置密码结构体
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// RefreshTokenRequest 刷新令牌请求结构体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户，默认获得user角色
// @Tags 认证
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "用户注册信息（默认user角色）"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/auth/register [post]
func (h *AuthController) Register(c *gin.Context) {
	if !h.cfg.SystemSettings.General.RegisterEnabled {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrRegisterDisabled)
		c.JSON(statusCode, errorResponse)
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	message, err := h.authService.Register(req.Username, req.Password, req.Email, req.Locale)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse(message))
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param user body LoginRequest true "用户登录信息"
// @Success 200 {object} success.DataResponse{data=services.LoginResponse} "登录成功，返回用户信息和访问令牌和刷新令牌"
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	loginResponse, err := h.authService.Login(req.Username, req.Password, &h.cfg.JWT)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Login successful", loginResponse))
}

// RequestPasswordReset 请求密码重置
// @Summary 请求密码重置
// @Description 发送密码重置邮件
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RequestPasswordResetRequest true "邮箱地址"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/auth/reset-password/request [post]
func (h *AuthController) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	err := h.authService.RequestPasswordReset(req.Email, req.Locale)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Password reset email sent"))
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用验证码设置新密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "邮箱、验证码和新密码"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/auth/reset-password [post]
func (h *AuthController) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	err := h.authService.ResetPassword(req.Email, req.Code, req.NewPassword)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Password reset successfully"))
}

// VerifyEmail 邮箱验证
// @Summary 邮箱验证
// @Description 使用验证令牌激活用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param token query string true "验证令牌"
// @Success 200 {object} success.SuccessResponse
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/auth/verify-email [get]
func (h *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.Data(statusCode, "text/html; charset=utf-8", []byte(fmt.Sprintf(errorHTML, errorResponse.Message)))
		return
	}

	err := h.authService.VerifyEmail(token)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.Data(statusCode, "text/html; charset=utf-8", []byte(fmt.Sprintf(errorHTML, errorResponse.Message)))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(successHTML, "Email verified successfully")))
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} success.DataResponse{data=services.TokenResponse} "刷新成功，返回新的访问令牌和刷新令牌"
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 403 {object} cerrors.ErrorResponse
// @Failure 404 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/auth/refresh [post]
func (h *AuthController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	tokenResponse, err := h.authService.RefreshToken(req.RefreshToken, &h.cfg.JWT)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewDataResponse("Token refreshed successfully", tokenResponse))
}

// Logout 用户登出
// @Summary 用户登出
// @Description 将刷新令牌添加到黑名单，防止再次使用
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "刷新令牌"
// @Success 200 {object} success.SuccessResponse "Logout successful"
// @Failure 400 {object} cerrors.ErrorResponse
// @Failure 401 {object} cerrors.ErrorResponse
// @Failure 500 {object} cerrors.ErrorResponse
// @Router /api/auth/logout [post]
func (h *AuthController) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrBadRequest)
		c.JSON(statusCode, errorResponse)
		return
	}

	err := h.authService.Logout(req.RefreshToken, &h.cfg.JWT)
	if err != nil {
		statusCode, errorResponse := cerrors.NewErrorResponse(err)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, success.NewSuccessResponse("Logout successful"))
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Produce json
// @Success 200 {object} success.DataResponse "data={"status": "ok"}"
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, success.NewDataResponse("Service is healthy", gin.H{"status": "ok"}))
}
