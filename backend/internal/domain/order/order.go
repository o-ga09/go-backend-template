package order

import (
	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/pkg/authz"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// IsOwnedBy はrequesterID(ログイン中ユーザーのID)がこの注文リソース本人か
// どうかを判定する(認可)。
func (o *Order) IsOwnedBy(requesterID string) bool {
	return authz.IsOwner(requesterID, o.UserID)
}

// NewFromCart はカートと、カート内商品IDをキーとした商品情報からOrderを組み立てる
// (architecture.mdの「ドメイン間に新しい矢印を作らない」方針に従い、orderパッケージが
// cart/productパッケージの型をimportして判定に使うのは許容されるが、cart/product側は
// orderをimportしない一方向の依存とする)。
//
// カート内の各商品について products[item.ProductID] の在庫を product.HasStock で
// 判定し、不足があればerrors.ErrInsufficientStockを返す。呼び出し元がこれを
// errors.MakeBusinessErrorに変換する。products に対応する商品が無い場合は
// errors.ErrRecordNotFoundを返す(呼び出し元は事前にproductRepo.FindByIDで存在確認
// している前提のため、通常到達しない防御的な分岐)。
func NewFromCart(c *cart.Cart, products map[string]*product.Product) (*Order, error) {
	items := make([]OrderItem, 0, len(c.Items))
	total := 0
	for _, item := range c.Items {
		p, ok := products[item.ProductID]
		if !ok {
			return nil, errors.ErrRecordNotFound
		}
		if !p.HasStock(item.Quantity) {
			return nil, errors.ErrInsufficientStock
		}

		items = append(items, OrderItem{
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			UnitPriceYen: p.PriceYen,
		})
		total += p.PriceYen * item.Quantity
	}

	return &Order{
		UserID:        c.UserID,
		Status:        StatusPending,
		TotalPriceYen: total,
		Items:         items,
	}, nil
}
