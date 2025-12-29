package router

import (
	"github.com/gin-gonic/gin"

	"isrvd/server/config"
	"isrvd/server/handler"
	"isrvd/server/middleware"
)

// 设置路由
func Setup() *gin.Engine {
	r := gin.Default()

	// 设置 CORS 中间件
	r.Use(middleware.CORS())

	// 创建处理器实例
	authHandler := handler.NewAuthHandler()
	fileHandler := handler.NewFileHandler()
	shellHandler := handler.NewShellHandler()
	zipHandler := handler.NewZipHandler()

	// API 路由组
	api := r.Group("/api")
	{
		// 公开路由
		api.POST("/login", authHandler.Login)

		// 需认证的路由组
		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", authHandler.Me)
			auth.POST("/list", fileHandler.List)
			auth.POST("/download", fileHandler.Download)
			auth.POST("/read", fileHandler.Read)

			// 管理员权限路由
			admin := auth.Group("")
			admin.Use(middleware.RequireRole(config.RoleAdmin))
			{
				admin.POST("/upload", fileHandler.Upload)
				admin.POST("/delete", fileHandler.Delete)
				admin.POST("/mkdir", fileHandler.Mkdir)
				admin.POST("/create", fileHandler.Create)
				admin.POST("/modify", fileHandler.Modify)
				admin.POST("/rename", fileHandler.Rename)
				admin.POST("/chmod", fileHandler.Chmod)
				admin.POST("/zip", zipHandler.Zip)
				admin.POST("/unzip", zipHandler.Unzip)
			}
		}
	}

	// WebSocket 路由 (仅管理员)
	r.GET("/ws/shell", middleware.Auth(), middleware.RequireRole(config.RoleAdmin), shellHandler.HandleWebSocket)

	return r
}
