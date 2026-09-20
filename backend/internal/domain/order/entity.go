// Package order は注文ドメイン(エンティティ + リポジトリinterface)を定義する。
// カートから注文を確定するAPI・注文履歴取得APIは認証必須で提供される
// (backend/tmp/task-4-brief.md参照)。
package order

import (
	"context"

	"github.com/o-ga09/go-backend-template/pkg/model"
)

//go:generate go run github.com/matryer/moq@latest -out mock/order_repository_mock.go -pkg moq . IOrderRepository

// StatusPending は注文の初期ステータス。状態遷移(発送・キャンセル等)は
// 本Issue(#4)のスコープ外のため固定値として扱う(YAGNI)。
const StatusPending = "pending"

// Order は注文ドメインエンティティ。domain＝DBモデルの方針に従い、
// GORMモデルを兼ねる(internal/database/mysql.orderRepository経由で永続化される)。
type Order struct {
	model.BaseModel
	UserID        string      `gorm:"column:user_id"`
	Status        string      `gorm:"column:status"`
	TotalPriceYen int         `gorm:"column:total_price_yen"`
	Items         []OrderItem `gorm:"-"` // GORMのhas many関連付けは使わず、リポジトリ層で個別に取得・詰め替える
}

// TableName はGORMが使用するテーブル名を明示する。
func (Order) TableName() string {
	return "orders"
}

// OrderItem は注文内の商品明細エンティティ。注文確定時点の単価を保持する
// (商品の価格が後で変更されても、過去の注文の金額が変わらないようにするため)。
type OrderItem struct {
	model.BaseModel
	OrderID      string `gorm:"column:order_id"`
	ProductID    string `gorm:"column:product_id"`
	Quantity     int    `gorm:"column:quantity"`
	UnitPriceYen int    `gorm:"column:unit_price_yen"`
}

// TableName はGORMが使用するテーブル名を明示する。
func (OrderItem) TableName() string {
	return "order_items"
}

// IOrderRepository は注文ドメインの永続化用インターフェース。
// 実装はinternal/database/mysql.orderRepository(GORM)。
type IOrderRepository interface {
	// Create は注文を新規作成する。o.Itemsもあわせてorder_itemsへ書き込む。
	// トランザクション制御は呼び出し元(ハンドラ)がITransactionManager.RunInTxで
	// 行う前提で、このメソッド自体はトランザクションを意識しない。
	Create(ctx context.Context, o *Order) error
	// FindByID はIDで注文(明細含む)を検索する。
	// 見つからない場合はerrors.ErrRecordNotFoundを返す。
	FindByID(ctx context.Context, id string) (*Order, error)
	// ListByUserID はユーザーIDで注文一覧(明細含む)を検索する。
	// ページングは今回不要(YAGNI)。
	ListByUserID(ctx context.Context, userID string) ([]*Order, error)
}
