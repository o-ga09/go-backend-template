package response

import "github.com/o-ga09/go-backend-template/internal/domain/order"

type OrderItem struct {
	ProductID    string `json:"productId"`
	Quantity     int    `json:"quantity"`
	UnitPriceYen int    `json:"unitPriceYen"`
}

type Order struct {
	ID            string      `json:"id"`
	Status        string      `json:"status"`
	TotalPriceYen int         `json:"totalPriceYen"`
	Items         []OrderItem `json:"items"`
}

func FromOrder(o *order.Order) Order {
	items := make([]OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, OrderItem{
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			UnitPriceYen: item.UnitPriceYen,
		})
	}
	return Order{
		ID:            o.ID,
		Status:        o.Status,
		TotalPriceYen: o.TotalPriceYen,
		Items:         items,
	}
}

func FromOrders(os []*order.Order) []Order {
	res := make([]Order, 0, len(os))
	for _, o := range os {
		res = append(res, FromOrder(o))
	}
	return res
}
