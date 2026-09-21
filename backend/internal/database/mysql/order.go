package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/internal/domain/order"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
)

// リポジトリはデータの読み書きのみを行う(在庫チェック・注文の組み立てはハンドラ層/
// domainの純粋関数で行う。architecture.mdの依存方向を厳守)。
type orderRepository struct{}

func NewOrderRepository() order.IOrderRepository {
	return &orderRepository{}
}

// o.ItemsはGORMのhas manyアソシエーションのため、db.Create(o)がorder_itemsも
// あわせて1回のバッチINSERTで書き込む(BaseModelPluginがスライスのID/Version採番に
// 対応しているため成立する。base_model_plugin.go参照)。
func (r *orderRepository) Create(ctx context.Context, o *order.Order) error {
	return Ctx.GetDBFromCtx(ctx).Create(o).Error
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (*order.Order, error) {
	var o order.Order
	if err := Ctx.GetDBFromCtx(ctx).Preload("Items", orderItemsAsc).Where("id = ?", id).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

// ページングは今回不要(YAGNI)。
func (r *orderRepository) ListByUserID(ctx context.Context, userID string) ([]*order.Order, error) {
	var orders []*order.Order
	if err := Ctx.GetDBFromCtx(ctx).Preload("Items", orderItemsAsc).Where("user_id = ?", userID).Order("created_at DESC, id DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func orderItemsAsc(db *gorm.DB) *gorm.DB {
	return db.Order("created_at ASC, id ASC")
}
