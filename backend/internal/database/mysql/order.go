package mysql

import (
	"context"
	stderrors "errors"

	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/internal/domain/order"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

// orderRepository はorderドメインのGORMリポジトリ。
// *gorm.DBをフィールドに保持せず、呼び出しごとにctxから取得する(user.goと同じ理由)。
// リポジトリはデータの読み書きのみを行い、ビジネスロジックを持たない
// (在庫チェック・注文の組み立てはハンドラ層/domainの純粋関数で行う。
// architecture.mdの依存方向を厳守)。
type orderRepository struct{}

// NewOrderRepository はorderRepositoryを生成する。
func NewOrderRepository() order.IOrderRepository {
	return &orderRepository{}
}

// Create はordersを1行と、o.Itemsの件数分order_itemsをINSERTする。
// このメソッド自体はトランザクションを意識しない。呼び出し元(ハンドラ)が
// ITransactionManager.RunInTxで包む前提で、ctxから取得した*gorm.DBをそのまま
// 使うだけでよい(transaction.md)。
//
// order_itemsはBaseModelPluginが単一構造体のreflect操作を前提としているため、
// バッチINSERT(スライスまとめて渡す)ではなく1件ずつCreateする
// (cart.goのAddItem等、既存リポジトリと同じ単発Create方式に合わせる)。
func (r *orderRepository) Create(ctx context.Context, o *order.Order) error {
	db := Ctx.GetDBFromCtx(ctx)

	if err := db.WithContext(ctx).Create(o).Error; err != nil {
		return err
	}

	for i := range o.Items {
		o.Items[i].OrderID = o.ID
		if err := db.WithContext(ctx).Create(&o.Items[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// FindByID はIDで注文(明細含む)を検索する。
func (r *orderRepository) FindByID(ctx context.Context, id string) (*order.Order, error) {
	db := Ctx.GetDBFromCtx(ctx)

	var o order.Order
	if err := db.WithContext(ctx).Where("id = ?", id).First(&o).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrRecordNotFound
		}
		return nil, err
	}

	items, err := r.findItems(ctx, db, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

// findItems は注文IDに紐づく明細を作成日時の昇順で取得する。
func (r *orderRepository) findItems(ctx context.Context, db *gorm.DB, orderID string) ([]order.OrderItem, error) {
	var items []order.OrderItem
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ListByUserID はユーザーIDで注文一覧(明細含む)を作成日時の降順で取得する。
// ページングは今回不要(YAGNI)。
func (r *orderRepository) ListByUserID(ctx context.Context, userID string) ([]*order.Order, error) {
	db := Ctx.GetDBFromCtx(ctx)

	var orders []*order.Order
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC, id DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	for _, o := range orders {
		items, err := r.findItems(ctx, db, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
	}
	return orders, nil
}
