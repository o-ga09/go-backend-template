package session_test

import (
	"strings"
	"testing"
	"time"

	"github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/session"
)

func TestManager_IssueAndVerify(t *testing.T) {
	mgr := session.NewManager("test-secret", time.Hour)

	token, expiresAt := mgr.Issue("user-1")
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatal("expiresAt should be in the future")
	}

	userID, err := mgr.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
	if userID != "user-1" {
		t.Errorf("Verify() userID = %q, want %q", userID, "user-1")
	}
}

func TestManager_Verify(t *testing.T) {
	mgr := session.NewManager("test-secret", time.Hour)
	otherMgr := session.NewManager("other-secret", time.Hour)
	expiredMgr := session.NewManager("test-secret", -time.Hour)

	validToken, _ := mgr.Issue("user-1")
	firstSegment, _, _ := strings.Cut(validToken, ".")
	expiredToken, _ := expiredMgr.Issue("user-1")
	wrongSigToken, _ := otherMgr.Issue("user-1")

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{"正常なトークンは検証できる", validToken, nil},
		{"フォーマットが不正なトークンは拒否される", "invalid-token", errors.ErrInvalidSession},
		{"パートが欠けたトークンは拒否される", firstSegment, errors.ErrInvalidSession},
		{"署名が異なるシークレットで発行されたトークンは拒否される", wrongSigToken, errors.ErrInvalidSession},
		{"有効期限切れのトークンは拒否される", expiredToken, errors.ErrSessionExpired},
		{"空文字のトークンは拒否される", "", errors.ErrInvalidSession},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mgr.Verify(tc.token)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Verify(%q) error = %v, want %v", tc.token, err, tc.wantErr)
			}
		})
	}
}
