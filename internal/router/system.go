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

	// DBヘルスチェック
	system.GET("/health/db", func(c echo.Context) error {
		db := Ctx.GetDBFromCtx(c.Request().Context())
		sqlDB, err := db.DB()
		if err != nil {
			return c.JSON(500, map[string]string{"status": "db connection error"})
		}
		if err := sqlDB.Ping(); err != nil {
			return c.JSON(500, map[string]string{"status": "db ping error"})
		}
		return c.JSON(200, map[string]string{"status": "db ok"})
	})

	// 時刻取得
	system.GET("/time", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"time": Ctx.GetRequestTime(c.Request().Context()).String()})
	})
}
