package user

import (
	"strings"

	"github.com/o-ga09/go-backend-template/pkg/authz"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// CanCreate はユーザー作成リクエストの入力が妥当かどうかを判定する
// (バリデーションはドメイン層の純粋関数として実装する。architecture.md参照)。
// 呼び出し元(handler)はこのエラーをerrors.MakeBusinessErrorに変換する。
func CanCreate(googleSub, displayName string) error {
	if strings.TrimSpace(googleSub) == "" {
		return errors.New("google sub is required")
	}
	if strings.TrimSpace(displayName) == "" {
		return errors.New("display name is required")
	}
	return nil
}

// IsOwnedBy はrequesterID(ログイン中ユーザーのID)がこのユーザーリソース本人か
// どうかを判定する(認可)。将来ドメイン(#4のcart/order等)も、
// リソースの所有者ID(例: order.UserID)とrequesterIDをpkg/authz.IsOwnerに渡す
// 同じ形で認可ロジックを再利用できる。
func (u *User) IsOwnedBy(requesterID string) bool {
	return authz.IsOwner(requesterID, u.ID)
}
