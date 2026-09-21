package response

import "github.com/o-ga09/go-backend-template/internal/domain/cart"

type CartItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type Cart struct {
	ID    string     `json:"id"`
	Items []CartItem `json:"items"`
}

func FromCart(c *cart.Cart) Cart {
	items := make([]CartItem, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, CartItem{ProductID: item.ProductID, Quantity: item.Quantity})
	}
	return Cart{ID: c.ID, Items: items}
}
