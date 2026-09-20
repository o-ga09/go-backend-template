package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
	moq "github.com/o-ga09/go-backend-template/internal/domain/user/mock"
	"github.com/o-ga09/go-backend-template/internal/handler"
	"github.com/o-ga09/go-backend-template/internal/server"
	"github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/session"
)

func TestAuthHandler_CurrentUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		repo       *moq.IUserRepositoryMock
		wantStatus int
	}{
		{
			name:   "セッションが有効ならログイン中ユーザー情報を返す",
			userID: "user-1",
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
			name:       "セッションが無い場合は401",
			userID:     "",
			repo:       &moq.IUserRepositoryMock{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "セッションが指すユーザーが存在しない場合も401",
			userID: "user-deleted",
			repo: &moq.IUserRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*user.User, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, "/api/auth/user", "", tc.userID)
			h := handler.NewAuthHandler(tc.repo)

			err := h.CurrentUser(c)
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

func TestAuthHandler_Logout(t *testing.T) {
	c, rec := newTestContext(t, http.MethodPost, "/api/auth/logout", "", "user-1")
	h := handler.NewAuthHandler(&moq.IUserRepositoryMock{})

	if err := h.Logout(c); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	cookie := findCookie(rec.Result().Cookies(), session.CookieName)
	if cookie == nil {
		t.Fatal("session cookie should be present (expired)")
		return
	}
	if cookie.MaxAge >= 0 {
		t.Errorf("cookie.MaxAge = %d, want negative (expired)", cookie.MaxAge)
	}
}
