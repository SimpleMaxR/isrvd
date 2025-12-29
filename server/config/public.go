package config

import (
	"embed"
)

// 静态文件
var PublicFS embed.FS

// 监听地址
var ListenAddr = ":8080"

// 基础目录
var BaseDirectory = "."

// 角色定义
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// 用户结构
type User struct {
	Password string
	Role     Role
}

// 用户名:用户对象
var Administrators = map[string]User{
	"admin": {Password: "admin", Role: RoleAdmin},
}
