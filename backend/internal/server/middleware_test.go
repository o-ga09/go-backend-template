package server_test

import (
	"net/http"
	"testing"

	"github.com/o-ga09/go-backend-template/internal/server"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

// TestErrCodeToStatusAndMessage_AuthErrors は、認証失敗(401)と認可失敗(403)の
// エラーコードが入れ替わっていた既存バグに対する回帰テスト。
// MakeAuthorizedError(認証失敗)は401、MakeAuthorizationError(認可失敗)は403に
// マッピングされなければならない(.claude/rules/error-handling.mdの表を参照)。
func TestErrCodeToStatusAndMessage_AuthErrors(t *testing.T) {
	ctx := Ctx.SetRequestID(t.Context(), uuid.GenerateID())

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"MakeAuthorizedErrorは401(認証失敗)", errors.MakeAuthorizedError(ctx, "invalid session"), http.StatusUnauthorized},
		{"MakeAuthorizationErrorは403(認可失敗)", errors.MakeAuthorizationError(ctx, "cannot access other user's resource"), http.StatusForbidden},
		{"MakeNotFoundErrorは404", errors.MakeNotFoundError(ctx, "not found"), http.StatusNotFound},
		{"MakeConflictErrorは409", errors.MakeConflictError(ctx, "conflict"), http.StatusConflict},
		{"MakeBusinessErrorは422", errors.MakeBusinessError(ctx, "invalid"), http.StatusUnprocessableEntity},
		{"MakeSystemErrorは500", errors.MakeSystemError(ctx, errors.New("boom")), http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, _ := server.ErrCodeToStatusAndMessage(tc.err)
			if gotStatus != tc.wantStatus {
				t.Errorf("ErrCodeToStatusAndMessage() status = %d, want %d", gotStatus, tc.wantStatus)
			}
		})
	}
}
