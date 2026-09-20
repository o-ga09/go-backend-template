package cart

import (
	"github.com/o-ga09/go-backend-template/pkg/authz"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// IsOwnedBy はrequesterID(ログイン中ユーザーのID)がこのカートリソース本人か
// どうかを判定する(認可)。
func (c *Cart) IsOwnedBy(requesterID string) bool {
	return authz.IsOwner(requesterID, c.UserID)
}

// CanCheckout はカートが注文確定可能な状態かどうかを判定する
// (バリデーションはドメイン層の純粋関数として実装する。architecture.md参照)。
// アイテムが空の場合は注文できない。呼び出し元(Task 4のハンドラ/service)は
// このエラーをerrors.MakeBusinessErrorに変換する。
func (c *Cart) CanCheckout() error {
	if len(c.Items) == 0 {
		return errors.New("cart has no items")
	}
	return nil
}
