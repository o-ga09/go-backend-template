package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
	moq "github.com/o-ga09/go-backend-template/internal/domain/user/mock"
	"github.com/o-ga09/go-backend-template/internal/handler"
	"github.com/o-ga09/go-backend-template/internal/server"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/session"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
	"github.com/o-ga09/go-backend-template/pkg/validator"
)

// newTestContext はerrors.Make*Error(内部でlogger経由でRequestIDを参照する)を
// panicさせないよう、RequestIDを設定したcontextを積んだecho.Contextを作る。
func newTestContext(t *testing.T, method, path string, body string, userID string) (*echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	e.Validator = validator.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	ctx := Ctx.SetRequestID(t.Context(), uuid.GenerateID())
	if userID != "" {
		ctx = Ctx.SetUserID(ctx, userID)
	}
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestUserHandler_Create(t *testing.T) {
	sessionMgr := session.NewManager("test-secret", time.Hour)

	tests := []struct {
		name       string
		body       string
		repo       *moq.IUserRepositoryMock
		wantStatus int
	}{
		{
			name: "新規ユーザーは作成されセッションCookieが発行される",
			body: `{"uid":"google-sub-1","displayName":"taro","profileImage":"https://example.com/a.png"}`,
			repo: &moq.IUserRepositoryMock{
				FindByGoogleSubFunc: func(ctx context.Context, googleSub string) (*user.User, error) {
					return nil, errors.ErrRecordNotFound
				},
				CreateFunc: func(ctx context.Context, u *user.User) error {
					u.ID = "user-1"
					return nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "既存ユーザーは特定されて200を返す",
			body: `{"uid":"google-sub-1","displayName":"taro","profileImage":"https://example.com/a.png"}`,
			repo: &moq.IUserRepositoryMock{
				FindByGoogleSubFunc: func(ctx context.Context, googleSub string) (*user.User, error) {
					return &user.User{GoogleSub: googleSub, Name: "taro"}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "uidが空の場合は422",
			body:       `{"uid":"","displayName":"taro"}`,
			repo:       &moq.IUserRepositoryMock{},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodPost, "/api/users", tc.body, "")
			h := handler.NewUserHandler(tc.repo, sessionMgr)

			err := h.Create(c)

			if err != nil {
				httpStatus, _ := server.ErrCodeToStatusAndMessage(err)
				if httpStatus != tc.wantStatus {
					t.Fatalf("error status = %d, want %d (err=%v)", httpStatus, tc.wantStatus, err)
				}
				return
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantStatus < http.StatusBadRequest {
				if cookie := findCookie(rec.Result().Cookies(), session.CookieName); cookie == nil {
					t.Error("session cookie should be set")
				}
			}
		})
	}
}

func TestUserHandler_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		targetID    string
		repo        *moq.IUserRepositoryMock
		wantStatus  int
	}{
		{
			name:        "本人のリソースは取得できる",
			requesterID: "user-1",
			targetID:    "user-1",
			repo: &moq.IUserRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*user.User, error) {
					u := &user.User{Name: "taro"}
					u.ID = id
					return u, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "他人のリソースは403で拒否される",
			requesterID: "user-1",
			targetID:    "user-2",
			repo: &moq.IUserRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*user.User, error) {
					u := &user.User{Name: "jiro"}
					u.ID = id
					return u, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:        "未ログインの場合は401",
			requesterID: "",
			targetID:    "user-2",
			repo:        &moq.IUserRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "存在しないユーザーは404",
			requesterID: "user-1",
			targetID:    "user-404",
			repo: &moq.IUserRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*user.User, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, "/api/users/:id", "", tc.requesterID)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tc.targetID}})

			h := handler.NewUserHandler(tc.repo, session.NewManager("secret", time.Hour))
			err := h.GetByID(c)

			if err != nil {
				status, _ := server.ErrCodeToStatusAndMessage(err)
				if status != tc.wantStatus {
					t.Fatalf("error status = %d, want %d (err=%v)", status, tc.wantStatus, err)
				}
				return
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}
