package user_test

import (
	"testing"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
)

func TestCanCreate(t *testing.T) {
	tests := []struct {
		name        string
		googleSub   string
		displayName string
		wantErr     bool
	}{
		{"google subと表示名が揃っていれば作成できる", "google-sub-1", "taro", false},
		{"google subが空の場合は作成できない", "", "taro", true},
		{"google subが空白のみの場合は作成できない", "   ", "taro", true},
		{"表示名が空の場合は作成できない", "google-sub-1", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := user.CanCreate(tc.googleSub, tc.displayName)
			if tc.wantErr {
				if err == nil {
					t.Fatal("error should not be nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("error should be nil, got %v", err)
			}
		})
	}
}

func TestUser_IsOwnedBy(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		requesterID string
		want        bool
	}{
		{"本人からのアクセスは許可される", "user-1", "user-1", true},
		{"他人からのアクセスは拒否される", "user-1", "user-2", false},
		{"未ログイン(空文字)からのアクセスは拒否される", "user-1", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &user.User{}
			u.ID = tc.userID

			got := u.IsOwnedBy(tc.requesterID)
			if got != tc.want {
				t.Errorf("IsOwnedBy(%q) = %v, want %v", tc.requesterID, got, tc.want)
			}
		})
	}
}
