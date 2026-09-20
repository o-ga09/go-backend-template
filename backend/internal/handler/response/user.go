// Package response はハンドラが返すHTTPレスポンスボディの型を定義する。
package response

import "github.com/o-ga09/go-backend-template/internal/domain/user"

// User はクライアント(frontend/api/user/types.ts のUser型)に返すユーザー情報。
// ProfileImageは現時点でusersテーブルに保存先のカラムが無いため常に空文字を返す
// (backend/docs/auth.md「既知の制限」参照)。
type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	ProfileImage string `json:"profileImage"`
}

// FromUser はdomain.UserからレスポンスのUserを組み立てる。
func FromUser(u *user.User) User {
	return User{
		ID:           u.ID,
		Name:         u.Name,
		DisplayName:  u.Name,
		ProfileImage: "",
	}
}
