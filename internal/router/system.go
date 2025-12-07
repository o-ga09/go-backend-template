package router

import (
	"github.com/labstack/echo"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
)

func SetupSystemRoute(root *echo.Group) {
	system := root.Group("/system")

	// ヘルスチェック
	system.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// 時刻取得
	system.GET("/time", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"time": Ctx.GetRequestTime(c.Request().Context()).String()})
	})
}
