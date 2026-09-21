package order

import (
	"context"

	"github.com/o-ga09/go-backend-template/pkg/model"
)

//go:generate go run github.com/matryer/moq@latest -out mock/order_repository_mock.go -pkg moq . IOrderRepository

// 状態遷移(発送・キャンセル等)は本Issue(#4)のスコープ外のため固定値として扱う(YAGNI)。
const StatusPending = "pending"

type Order struct {
	model.BaseModel
	UserID        string      `gorm:"column:user_id"`
	Status        string      `gorm:"column:status"`
	TotalPriceYen int         `gorm:"column:total_price_yen"`
	Items         []OrderItem `gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return "orders"
}

// 注文確定時点の単価を保持する(商品の価格が後で変更されても過去の注文の金額が
// 変わらないようにするため)。
type OrderItem struct {
	model.BaseModel
	OrderID      string `gorm:"column:order_id"`
	ProductID    string `gorm:"column:product_id"`
	Quantity     int    `gorm:"column:quantity"`
	UnitPriceYen int    `gorm:"column:unit_price_yen"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

type IOrderRepository interface {
	// トランザクション制御は呼び出し元(ハンドラ)がITransactionManager.RunInTxで
	// 行う前提で、このメソッド自体はトランザクションを意識しない。
	Create(ctx context.Context, o *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	ListByUserID(ctx context.Context, userID string) ([]*Order, error)
}
