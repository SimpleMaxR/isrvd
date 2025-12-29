package service

import (
	"errors"
	"sync"
	"time"

	"isrvd/server/config"
	"isrvd/server/helper"
	"isrvd/server/model"
)

// 认证服务
type Session struct {
	Username string
	Role     config.Role
	Expiry   time.Time
}

type AuthService struct {
	mutex    sync.RWMutex
	sessions map[string]Session
}

// 认证服务实例
var AuthInstance *AuthService

// 创建认证服务实例
func GetAuthService() *AuthService {
	if AuthInstance == nil {
		AuthInstance = &AuthService{
			sessions: make(map[string]Session),
		}
		go AuthInstance.CleanupExpired()
	}
	return AuthInstance
}

// 用户登录
func (as *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	// 验证用户名和密码
	if user, exists := config.Administrators[req.Username]; exists && user.Password == req.Password {
		return &model.LoginResponse{
			Token:    as.CreateToken(req.Username, user.Role),
			Username: req.Username,
			Role:     string(user.Role),
		}, nil
	}

	return nil, errors.New("invalid credentials")
}

// 创建令牌
func (as *AuthService) CreateToken(username string, role config.Role) string {
	token := helper.Md5sum(username + time.Now().String())
	as.mutex.Lock()
	as.sessions[token] = Session{
		Username: username,
		Role:     role,
		Expiry:   time.Now().Add(24 * time.Hour),
	}
	as.mutex.Unlock()
	return token
}

// 删除令牌
func (as *AuthService) DeleteToken(token string) {
	as.mutex.Lock()
	delete(as.sessions, token)
	as.mutex.Unlock()
}

// 验证令牌
func (as *AuthService) ValidateToken(token string) bool {
	as.mutex.RLock()
	session, exists := as.sessions[token]
	as.mutex.RUnlock()

	if !exists {
		return false
	}

	if session.Expiry.Before(time.Now()) {
		as.DeleteToken(token)
		return false
	}

	return true
}

// 获取会话信息
func (as *AuthService) GetSession(token string) (Session, bool) {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	session, exists := as.sessions[token]
	return session, exists
}

// 清理过期的会话
func (as *AuthService) CleanupExpired() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		as.mutex.Lock()
		now := time.Now()
		for token, session := range as.sessions {
			if session.Expiry.Before(now) {
				delete(as.sessions, token)
			}
		}
		as.mutex.Unlock()
	}
}
