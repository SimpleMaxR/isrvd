package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/server/helper"
	"isrvd/server/model"
	"isrvd/server/service"
)

// 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.GetAuthService(),
	}
}

// 登录处理
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.RespondError(c, http.StatusBadRequest, "Invalid JSON")
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		helper.RespondError(c, http.StatusUnauthorized, err.Error())
		return
	}

	resp.Permissions = h.getPermissions(resp.Role)

	helper.RespondSuccess(c, "Login successful", resp)
}

// 登出处理
func (h *AuthHandler) Logout(c *gin.Context) {
	token := helper.GetTokenFromRequest(c)
	h.authService.DeleteToken(token)

	helper.RespondSuccess(c, "Logged out successfully", nil)
}

// 获取当前用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	token := helper.GetTokenFromRequest(c)
	session, exists := h.authService.GetSession(token)
	if !exists {
		helper.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp := model.UserInfoResponse{
		Username:    session.Username,
		Role:        string(session.Role),
		Permissions: h.getPermissions(string(session.Role)),
	}

	helper.RespondSuccess(c, "Get user info successful", resp)
}

// 用户权限配置
func (h *AuthHandler) getPermissions(role string) model.Permissions {
	if role == "admin" {
		return model.Permissions{
			CanRead:    true,
			CanWrite:   true,
			CanDelete:  true,
			CanExecute: true,
		}
	}
	// 默认普通用户只有读权限
	return model.Permissions{
		CanRead:    true,
		CanWrite:   true,
		CanDelete:  true,
		CanExecute: false,
	}
}
