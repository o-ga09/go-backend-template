// Package user はユーザードメイン(エンティティ + リポジトリinterface)を定義する。
// NextAuth(Google OAuth)でログインしたクライアントが、バックエンド側で
// 識別/永続化されるユーザーを表す。
package user

import (
	"context"

	"github.com/o-ga09/go-backend-template/pkg/model"
)

//go:generate go run github.com/matryer/moq@latest -out mock/user_repository_mock.go -pkg moq . IUserRepository

// User はユーザードメインエンティティ。domain＝DBモデルの方針に従い、
// GORMモデルを兼ねる(internal/database/mysql.UserRepository経由で永続化される)。
type User struct {
	model.BaseModel
	GoogleSub string `gorm:"column:google_sub"`
	Email     string `gorm:"column:email"`
	Name      string `gorm:"column:name"`
}

// TableName はGORMが使用するテーブル名を明示する。
func (User) TableName() string {
	return "users"
}

// IUserRepository はユーザードメインの永続化用インターフェース。
// 実装はinternal/database/mysql.UserRepository(GORM)。
type IUserRepository interface {
	// FindByID はIDでユーザーを検索する。見つからない場合はerrors.ErrRecordNotFoundを返す。
	FindByID(ctx context.Context, id string) (*User, error)
	// FindByGoogleSub はGoogleのsubでユーザーを検索する。見つからない場合はerrors.ErrRecordNotFoundを返す。
	FindByGoogleSub(ctx context.Context, googleSub string) (*User, error)
	// Create はユーザーを新規作成する。ID/Version/タイムスタンプはGORMプラグインが採番する。
	Create(ctx context.Context, u *User) error
	// Update はユーザーを更新する(楽観ロック)。競合時はerrors.ErrOptimisticLockConflictを返す。
	Update(ctx context.Context, u *User) error
}
