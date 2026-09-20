package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/server"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/session"
)

func TestAuthenticate(t *testing.T) {
	mgr := session.NewManager("test-secret", time.Hour)
	validToken, _ := mgr.Issue("user-1")

	tests := []struct {
		name       string
		cookie     *http.Cookie
		wantUserID string
	}{
		{"有効なセッションCookieがあればユーザーIDがcontextに入る", &http.Cookie{Name: session.CookieName, Value: validToken}, "user-1"},
		{"Cookieが無い場合はユーザーIDが空のまま次に進む", nil, ""},
		{"不正なCookieの場合はユーザーIDが空のまま次に進む", &http.Cookie{Name: session.CookieName, Value: "invalid"}, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var gotUserID string
			handlerCalled := false
			next := func(c *echo.Context) error {
				handlerCalled = true
				gotUserID = Ctx.GetUserID(c.Request().Context())
				return nil
			}

			if err := server.Authenticate(mgr)(next)(c); err != nil {
				t.Fatalf("Authenticate() error = %v", err)
			}
			if !handlerCalled {
				t.Fatal("next handler should always be called (Authenticate never blocks the request)")
			}
			if gotUserID != tc.wantUserID {
				t.Errorf("userID = %q, want %q", gotUserID, tc.wantUserID)
			}
		})
	}
}
