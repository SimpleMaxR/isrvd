package model

// 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 登录响应结构
type LoginResponse struct {
	Token       string      `json:"token"`
	Username    string      `json:"username"`
	Role        string      `json:"role"`
	Permissions Permissions `json:"permissions"`
}

// 用户信息响应结构
type UserInfoResponse struct {
	Username    string      `json:"username"`
	Role        string      `json:"role"`
	Permissions Permissions `json:"permissions"`
}

// 权限结构
type Permissions struct {
	CanRead    bool `json:"canRead"`
	CanWrite   bool `json:"canWrite"`
	CanDelete  bool `json:"canDelete"`
	CanExecute bool `json:"canExecute"`
}
