package router

import (
	"github.com/labstack/echo/v5"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
)

func (r *route) SetupSystemRoute() {
	system := r.rooAPI.Group("/system")

	// ヘルスチェック
	system.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// DBヘルスチェック
	system.GET("/health/db", func(c *echo.Context) error {
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
	system.GET("/time", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"time": Ctx.GetRequestTime(c.Request().Context()).String()})
	})

	// CSRFトークン発行。internal/server/middleware.goのCSRFProtectionミドルウェアが
	// 検証に使うトークンをJSONで返す(ダブルサブミットCookie方式のフォールバック用)。
	// フロントエンドの正規オリジンからのfetch/XHRはSec-Fetch-Siteヘッダーにより
	// 自動的に許可されるため、通常はこのエンドポイントを呼ぶ必要はない。
	// Sec-Fetch-Siteを送らない環境向けに、このエンドポイントで取得したトークンを
	// 以降の状態変更リクエストのX-CSRF-Tokenヘッダーに設定する。
	r.rooAPI.GET("/csrf", func(c *echo.Context) error {
		token, _ := c.Get("csrf").(string)
		return c.JSON(200, map[string]string{"csrfToken": token})
	})
}
