package router

import (
	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/handler"
)

// SetupApplicationRoute はアプリケーション固有のAPIルーティングを登録する。
func SetupApplicationRoute(root *echo.Group, userHandler *handler.UserHandler, authHandler *handler.AuthHandler) {
	users := root.Group("/users")
	users.POST("", userHandler.Create)
	users.GET("/:id", userHandler.GetByID)

	auth := root.Group("/auth")
	auth.GET("/user", authHandler.CurrentUser)
	auth.POST("/logout", authHandler.Logout)
}
