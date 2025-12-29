package config

import (
	"os"
	"strings"
)

func init() {
	// 从环境变量读取端口
	if value := os.Getenv("LISTEN_ADDR"); value != "" {
		ListenAddr = value
	}

	// 从环境变量读取基础目录
	if value := os.Getenv("BASE_DIRECTORY"); value != "" {
		BaseDirectory = value
	}

	// 从环境变量读取用户配置
	if value := os.Getenv("ADMINISTRATORS"); value != "" {
		delete(Administrators, "admin") // 删除默认用户
		for _, pair := range strings.Split(value, ",") {
			parts := strings.Split(pair, ":")
			if len(parts) >= 2 {
				username := parts[0]
				password := parts[1]
				role := RoleUser // 默认为普通用户

				// 如果用户名为 admin，默认为管理员
				if username == "admin" {
					role = RoleAdmin
				}

				if len(parts) >= 3 {
					switch Role(parts[2]) {
					case RoleAdmin:
						role = RoleAdmin
					case RoleUser:
						role = RoleUser
					default:
						role = RoleUser // 未知角色降级为普通用户
					}
				}

				Administrators[username] = User{
					Password: password,
					Role:     role,
				}
			}
		}
	}
}
