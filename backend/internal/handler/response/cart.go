package response

import "github.com/o-ga09/go-backend-template/internal/domain/cart"

// CartItem はクライアントに返すカート内商品明細情報。
type CartItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// Cart はクライアントに返すカート情報。
type Cart struct {
	ID    string     `json:"id"`
	Items []CartItem `json:"items"`
}

// FromCart はdomain.CartからレスポンスのCartを組み立てる。
func FromCart(c *cart.Cart) Cart {
	items := make([]CartItem, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, CartItem{ProductID: item.ProductID, Quantity: item.Quantity})
	}
	return Cart{ID: c.ID, Items: items}
}
