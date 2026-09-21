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

// TestCSRFProtection_SecFetchSite は、フロントエンドの正規オリジンからの
// クロスオリジンリクエスト(Sec-Fetch-Siteヘッダー付き)がトークン無しで許可され、
// それ以外のクロスサイトオリジンは拒否されることを確認する。
func TestCSRFProtection_SecFetchSite(t *testing.T) {
	const trustedOrigin = "https://frontend.example.com"

	tests := []struct {
		name          string
		origin        string
		secFetchSite  string
		wantNextCalls int
		wantErr       bool
	}{
		{
			name:          "信頼済みオリジンからのクロスオリジンPOSTはトークン無しで許可される",
			origin:        trustedOrigin,
			secFetchSite:  "cross-site",
			wantNextCalls: 1,
		},
		{
			name:         "信頼されていないオリジンからのクロスサイトPOSTは拒否される",
			origin:       "https://evil.example.com",
			secFetchSite: "cross-site",
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Sec-Fetch-Site", tc.secFetchSite)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalls := 0
			next := func(c *echo.Context) error {
				nextCalls++
				return nil
			}

			err := server.CSRFProtection(csrfTestCtx(t, trustedOrigin))(next)(c)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
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

	t.Run("トークンが無い状態変更リクエストは拒否される", func(t *testing.T) {
		e := echo.New()
		req := newCSRFTestRequest(t, http.MethodPost, "/api/orders")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c *echo.Context) error { return nil }

		if err := server.CSRFProtection(csrfTestCtx(t, trustedOrigin))(next)(c); err == nil {
			t.Fatal("expected an error, got nil")
		}
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
}
