package order

import (
	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/pkg/authz"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

func (o *Order) IsOwnedBy(requesterID string) bool {
	return authz.IsOwner(requesterID, o.UserID)
}

// order側からcart/productを参照する一方向の依存にする(cart/product側はorderを
// 参照しない。architecture.md「ドメイン間に新しい矢印を作らない」)。
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
