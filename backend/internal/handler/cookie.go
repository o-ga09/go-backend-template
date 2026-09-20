package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/pkg/session"
)

// setSessionCookie はセッションCookieを発行し、レスポンスにSet-Cookieヘッダーを追加する。
// SameSite=None + Secure はフロントエンド(別オリジン)からcredentials: 'include'で
// アクセスするために必須(backend/docs/auth.md参照)。
func setSessionCookie(c *echo.Context, mgr *session.Manager, userID string) {
	token, expiresAt := mgr.Issue(userID)
	c.SetCookie(&http.Cookie{
		Name:     session.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

// clearSessionCookie はセッションCookieを失効させる(ログアウト)。
func clearSessionCookie(c *echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     session.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}
