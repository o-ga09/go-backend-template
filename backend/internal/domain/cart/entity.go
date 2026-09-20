// Package cart はカートドメイン(エンティティ + リポジトリinterface)を定義する。
// カートへの商品追加・数量変更・削除は認証必須のAPIとして提供される
// (backend/tmp/task-3-brief.md参照)。
package cart

import (
	"context"

	"github.com/o-ga09/go-backend-template/pkg/model"
)

//go:generate go run github.com/matryer/moq@latest -out mock/cart_repository_mock.go -pkg moq . ICartRepository

// Cart はカードメインエンティティ。domain＝DBモデルの方針に従い、
// GORMモデルを兼ねる(internal/database/mysql.cartRepository経由で永続化される)。
// 1ユーザー1カート(carts.user_idにUNIQUE制約)。
type Cart struct {
	model.BaseModel
	UserID string     `gorm:"column:user_id"`
	Items  []CartItem `gorm:"-"` // GORMのhas many関連付けは使わず、リポジトリ層で個別に取得・詰め替える
}

// TableName はGORMが使用するテーブル名を明示する。
func (Cart) TableName() string {
	return "carts"
}

// CartItem はカート内の商品明細エンティティ。
type CartItem struct {
	model.BaseModel
	CartID    string `gorm:"column:cart_id"`
	ProductID string `gorm:"column:product_id"`
	Quantity  int    `gorm:"column:quantity"`
}

// TableName はGORMが使用するテーブル名を明示する。
func (CartItem) TableName() string {
	return "cart_items"
}

// ICartRepository はカードメインの永続化用インターフェース。
// 実装はinternal/database/mysql.cartRepository(GORM)。
type ICartRepository interface {
	// FindByUserID はユーザーIDでカート(明細含む)を検索する。
	// 見つからない場合はerrors.ErrRecordNotFoundを返す。
	FindByUserID(ctx context.Context, userID string) (*Cart, error)
	// Create はユーザーのカートを新規作成する。ID/Version/タイムスタンプは
	// GORMプラグインが採番する。
	Create(ctx context.Context, userID string) (*Cart, error)
	// AddItem はカートに商品を追加する。既に同じ商品が入っている場合は
	// 数量を加算する(cart_items.uq_cart_items_cart_productのUNIQUE制約に対応)。
	AddItem(ctx context.Context, cartID, productID string, quantity int) error
	// UpdateItemQuantity はカート内の商品の数量を指定した値に変更する。
	// 対象の明細が無い場合はerrors.ErrRecordNotFoundを返す。
	UpdateItemQuantity(ctx context.Context, cartID, productID string, quantity int) error
	// RemoveItem はカートから商品を削除する。対象の明細が無い場合は
	// errors.ErrRecordNotFoundを返す。
	RemoveItem(ctx context.Context, cartID, productID string) error
	// Clear はカート内の全明細を削除する(注文確定時にTask 4から呼ばれる想定)。
	Clear(ctx context.Context, cartID string) error
}
