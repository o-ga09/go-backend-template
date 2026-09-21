package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/server"
	"github.com/o-ga09/go-backend-template/pkg/config"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

// csrfTestCtx はCSRFProtection(ctx)が読み取るpkg/config.ConfigをFrontendOriginだけ
// 設定したcontextに積んで返す(ミドルウェア構築時にのみ使う)。
func csrfTestCtx(t *testing.T, frontendOrigin string) context.Context {
	t.Helper()
	return context.WithValue(t.Context(), config.ConfigKey, &config.Config{FrontendOrigin: frontendOrigin})
}

// newCSRFTestRequest はhttptest.NewRequestに加え、リクエストのcontextへRequestIDを
// 積んだ*http.Requestを返す。CSRF拒否時にerrors.MakeAuthorizationErrorがログ出力の
// ためRequestIDを参照するため、実リクエストのcontext側にも必要になる
// (csrfTestCtxが返すcontextはミドルウェア構築用でリクエストには紐付かない)。
func newCSRFTestRequest(t *testing.T, method, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	ctx := Ctx.SetRequestID(t.Context(), uuid.GenerateID())
	return req.WithContext(ctx)
}

// assertForbidden は、CSRFProtection()が返したerrorが403(認可失敗)に変換される
// pkg/errors由来のエラーであることを検証する。echoのCSRFミドルウェアは拒否経路
// (Sec-Fetch-Site起因/ダブルサブミット起因)によって異なる型のerrorを返しうるため、
// 「errが非nil」だけでなく「ErrCodeToStatusAndMessageで実際に403になる」ことまで
// 確認する(500に落ちる回帰を検出するため)。
func assertForbidden(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	status, _ := server.ErrCodeToStatusAndMessage(err)
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (err=%v)", status, http.StatusForbidden, err)
	}
}

// TestCSRFProtection_SecFetchSite は、Sec-Fetch-Siteヘッダー(Fetch Metadata)による
// 検証の分岐を確認する。フロントエンドの正規オリジンからのクロスオリジンリクエストは
// トークン無しで許可され、それ以外は403(内部的にpkg/errors経由)になることを検証する。
func TestCSRFProtection_SecFetchSite(t *testing.T) {
	const trustedOrigin = "https://frontend.example.com"

	tests := []struct {
		name          string
		origin        string
		secFetchSite  string
		wantNextCalls int
		wantForbidden bool
	}{
		{
			name:          "信頼済みオリジンからのクロスオリジンPOSTはトークン無しで許可される",
			origin:        trustedOrigin,
			secFetchSite:  "cross-site",
			wantNextCalls: 1,
		},
		{
			name:          "信頼済みオリジンからのsame-siteのPOSTも許可される",
			origin:        trustedOrigin,
			secFetchSite:  "same-site",
			wantNextCalls: 1,
		},
		{
			name:          "same-originのPOSTは常に許可される",
			secFetchSite:  "same-origin",
			wantNextCalls: 1,
		},
		{
			name:          "直接ナビゲーション(none)は常に許可される",
			secFetchSite:  "none",
			wantNextCalls: 1,
		},
		{
			name:          "信頼されていないオリジンからのクロスサイトPOSTは403で拒否される",
			origin:        "https://evil.example.com",
			secFetchSite:  "cross-site",
			wantForbidden: true,
		},
		{
			name:          "信頼されていないオリジンからのsame-site偽装POSTも403で拒否される",
			origin:        "https://evil.example.com",
			secFetchSite:  "same-site",
			wantForbidden: true,
		},
		{
			name:          "Originが信頼済みでもSec-Fetch-Siteがcross-siteかつ一致しなければ拒否される(Origin単体は信頼しない)",
			origin:        trustedOrigin + ".evil.example.com",
			secFetchSite:  "cross-site",
			wantForbidden: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			req.Header.Set("Sec-Fetch-Site", tc.secFetchSite)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalls := 0
			next := func(c *echo.Context) error {
				nextCalls++
				return nil
			}

			err := server.CSRFProtection(csrfTestCtx(t, trustedOrigin))(next)(c)

			if tc.wantForbidden {
				assertForbidden(t, err)
				return
			}
			if err != nil {
				t.Fatalf("CSRFProtection() error = %v", err)
			}
			if nextCalls != tc.wantNextCalls {
				t.Errorf("next was called %d times, want %d", nextCalls, tc.wantNextCalls)
			}
		})
	}
}

// TestCSRFProtection_DoubleSubmitCookie は、Sec-Fetch-Siteヘッダーを送らない
// (古いブラウザ相当の)クライアントに対するダブルサブミットCookie方式の
// フォールバック動作を確認する。
func TestCSRFProtection_DoubleSubmitCookie(t *testing.T) {
	const trustedOrigin = "https://frontend.example.com"

	t.Run("トークンが無い状態変更リクエストは403で拒否される", func(t *testing.T) {
		e := echo.New()
		req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c *echo.Context) error { return nil }

		err := server.CSRFProtection(csrfTestCtx(t, trustedOrigin))(next)(c)
		assertForbidden(t, err)
	})

	t.Run("GETリクエストは常に許可される(safe method)", func(t *testing.T) {
		e := echo.New()
		req := newCSRFTestRequest(t, http.MethodGet, "/api/csrf")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		nextCalled := false
		next := func(c *echo.Context) error {
			nextCalled = true
			return nil
		}

		if err := server.CSRFProtection(csrfTestCtx(t, trustedOrigin))(next)(c); err != nil {
			t.Fatalf("CSRFProtection() error = %v", err)
		}
		if !nextCalled {
			t.Fatal("next handler should be called for GET requests")
		}
	})

	t.Run("GETで発行されたトークンをCookieとヘッダーの両方に付けたPOSTは許可される", func(t *testing.T) {
		mw := server.CSRFProtection(csrfTestCtx(t, trustedOrigin))

		// 1. まずGETでトークンを発行させる(_csrf Cookieがセットされる)。
		e := echo.New()
		getReq := newCSRFTestRequest(t, http.MethodGet, "/api/csrf")
		getRec := httptest.NewRecorder()
		getCtx := e.NewContext(getReq, getRec)
		if err := mw(func(c *echo.Context) error { return nil })(getCtx); err != nil {
			t.Fatalf("GET request error = %v", err)
		}

		var csrfCookie *http.Cookie
		for _, ck := range getRec.Result().Cookies() {
			if ck.Name == "_csrf" {
				csrfCookie = ck
			}
		}
		if csrfCookie == nil {
			t.Fatal("_csrf cookie was not set by the GET request")
		}
		if csrfCookie.Path != "/" {
			t.Errorf("_csrf cookie Path = %q, want %q (fallback breaks across different request paths otherwise)", csrfCookie.Path, "/")
		}

		// 2. 発行されたトークンをCookieとX-CSRF-Tokenヘッダーの両方に付けてPOSTする。
		postReq := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
		postReq.AddCookie(&http.Cookie{Name: "_csrf", Value: csrfCookie.Value})
		postReq.Header.Set(echo.HeaderXCSRFToken, csrfCookie.Value)
		postRec := httptest.NewRecorder()
		postCtx := e.NewContext(postReq, postRec)

		nextCalled := false
		if err := mw(func(c *echo.Context) error {
			nextCalled = true
			return nil
		})(postCtx); err != nil {
			t.Fatalf("POST request error = %v", err)
		}
		if !nextCalled {
			t.Fatal("next handler should be called when the CSRF token matches")
		}
	})

	t.Run("Cookieとヘッダーのトークンが一致しない場合は403で拒否される", func(t *testing.T) {
		mw := server.CSRFProtection(csrfTestCtx(t, trustedOrigin))

		req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
		req.AddCookie(&http.Cookie{Name: "_csrf", Value: "cookie-token"})
		req.Header.Set(echo.HeaderXCSRFToken, "different-header-token")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		err := mw(func(c *echo.Context) error { return nil })(c)
		assertForbidden(t, err)
	})
}

// TestCSRFProtection_WrappedByErrorHandler は、server.go でCSRFProtectionを
// ErrorHandler()より後(=ハンドラに近い内側)に登録することで、CSRF拒否が実際に
// ErrorHandler経由でpkg/errors形式の403レスポンスになることを、実際のミドルウェア
// チェーン(echo.New() + Use() + ServeHTTP)を通して確認する。ミドルウェア単体呼び出し
// だけでは登録順序に起因する不具合(例: 中間でechoの生HTTPErrorが素通しされ500に
// なる)を検出できないため、e2eに近い形で検証する。
func TestCSRFProtection_WrappedByErrorHandler(t *testing.T) {
	const trustedOrigin = "https://frontend.example.com"

	newEngine := func() *echo.Echo {
		e := echo.New()
		e.Use(server.ErrorHandler())
		e.Use(server.CSRFProtection(csrfTestCtx(t, trustedOrigin)))
		e.POST("/api/orders", func(c *echo.Context) error {
			return c.JSON(http.StatusCreated, map[string]string{"status": "ok"})
		})
		return e
	}

	t.Run("信頼されていないクロスサイトオリジンは403のJSONレスポンスになる(500に落ちない)", func(t *testing.T) {
		e := newEngine()
		req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
		req.Header.Set("Origin", "https://evil.example.com")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})

	t.Run("トークン無しのダブルサブミット経路も403のJSONレスポンスになる", func(t *testing.T) {
		e := newEngine()
		req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})

	t.Run("信頼済みオリジンからの正規リクエストはハンドラまで到達する", func(t *testing.T) {
		e := newEngine()
		req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
		req.Header.Set("Origin", trustedOrigin)
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
	})
}
